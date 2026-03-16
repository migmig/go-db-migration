package migration

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"dbmigrator/internal/bus"
	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
)

type GroupedMetadata struct {
	Tables    []string
	Sequences []dialect.SequenceMetadata
}

type GroupedScripts struct {
	Tables    []string
	Sequences []NamedScript
}

type NamedScript struct {
	Name string
	SQL  string
}

func collectGroupedMetadata(dbConn *sql.DB, cfg *config.Config) GroupedMetadata {
	grouped := GroupedMetadata{
		Tables: append([]string(nil), cfg.Tables...),
	}

	if !cfg.WithSequences || dbConn == nil {
		return grouped
	}

	owner := resolveOwner(cfg)
	seen := make(map[string]struct{})
	extraNames := splitNames(cfg.Sequences)

	for _, tableName := range cfg.Tables {
		seqs, err := GetSequenceMetadata(dbConn, tableName, owner, extraNames)
		if err != nil {
			slog.Warn("failed to get sequence metadata", "table", tableName, "error", err)
			continue
		}

		for _, seq := range seqs {
			if _, ok := seen[seq.Name]; ok {
				continue
			}
			seen[seq.Name] = struct{}{}
			grouped.Sequences = append(grouped.Sequences, seq)
		}
	}

	return grouped
}

func buildGroupedScripts(metadata GroupedMetadata, schema string, dia dialect.Dialect, tracker ProgressTracker) GroupedScripts {
	scripts := GroupedScripts{}

	for _, seq := range metadata.Sequences {
		ddl, supported := GenerateSequenceDDL(seq, schema, dia)
		if !supported || ddl == "" {
			slog.Warn("sequence not supported by dialect", "dialect", dia.Name(), "sequence", seq.Name)
			if tracker != nil {
				if tracker.EventBus() != nil {
					tracker.EventBus().Publish(bus.Event{
						Type:    bus.EventWarning,
						Message: fmt.Sprintf("%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.", dia.Name()),
					})
				} else if wt, ok := tracker.(WarningTracker); ok {
					wt.Warning(fmt.Sprintf("%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.", dia.Name()))
				}
			}
			continue
		}
		scripts.Sequences = append(scripts.Sequences, NamedScript{
			Name: seq.Name,
			SQL:  ddl,
		})
	}

	return scripts
}

func openSequenceWriter(cfg *config.Config, group string, mainBuf *bufio.Writer) (*bufio.Writer, func() error, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}

	if group == config.ObjectGroupSequences && !cfg.PerTable {
		if mainBuf == nil {
			return nil, nil, fmt.Errorf("output buffer is not initialized for sequences-only mode")
		}
		return mainBuf, func() error { return nil }, nil
	}

	fileName := "sequences.sql"
	if cfg.OutputDir != "" {
		fileName = cfg.OutputDir + "/" + fileName
	}

	flag := os.O_CREATE | os.O_WRONLY
	if cfg.ResumeJobID != "" {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	out, err := os.OpenFile(fileName, flag, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create sequences output file: %w", err)
	}

	buf := bufio.NewWriter(out)
	return buf, func() error {
		if err := buf.Flush(); err != nil {
			_ = out.Close()
			return err
		}
		return out.Close()
	}, nil
}

func migrateSequenceGroup(targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, cfg *config.Config, mainBuf *bufio.Writer, metadata GroupedMetadata, tracker ProgressTracker, report *MigrationReport) error {
	scripts := buildGroupedScripts(metadata, cfg.Schema, dia, tracker)
	skipped := len(metadata.Sequences) - len(scripts.Sequences)
	if report != nil && skipped > 0 {
		report.SkipGroup(config.ObjectGroupSequences, skipped)
	}

	if len(scripts.Sequences) == 0 {
		slog.Info("script.generated", "object_group", config.ObjectGroupSequences, "count", 0)
		return nil
	}

	var (
		seqBuf  *bufio.Writer
		closeFn func() error
		err     error
	)

	if pgPool == nil && targetDB == nil {
		seqBuf, closeFn, err = openSequenceWriter(cfg, cfg.ObjectGroup, mainBuf)
		if err != nil {
			return err
		}
		defer func() {
			if closeFn != nil {
				_ = closeFn()
			}
		}()
	}

	slog.Info("script.generated", "object_group", config.ObjectGroupSequences, "count", len(scripts.Sequences))

	ddlTracker, hasDDLTracker := tracker.(DDLProgressTracker)
	for _, seqScript := range scripts.Sequences {
		seqName := seqScript.Name

		switch {
		case pgPool != nil:
			if _, err := pgPool.Exec(context.Background(), seqScript.SQL); err != nil {
				if report != nil {
					report.RecordSequenceResult(err)
				}
				return publishSequenceError(tracker, ddlTracker, hasDDLTracker, seqName, err)
			}
		case targetDB != nil:
			if _, err := targetDB.Exec(seqScript.SQL); err != nil {
				if report != nil {
					report.RecordSequenceResult(err)
				}
				return publishSequenceError(tracker, ddlTracker, hasDDLTracker, seqName, err)
			}
		default:
			if _, err := seqBuf.WriteString(seqScript.SQL + "\n"); err != nil {
				if report != nil {
					report.RecordSequenceResult(err)
				}
				return publishSequenceError(tracker, ddlTracker, hasDDLTracker, seqName, err)
			}
		}

		if report != nil {
			report.RecordSequenceResult(nil)
		}
		if tracker != nil && tracker.EventBus() != nil {
			tracker.EventBus().Publish(bus.Event{
				Type:       bus.EventDDLProgress,
				Object:     "sequence",
				ObjectName: seqName,
				Status:     "ok",
			})
		} else if hasDDLTracker {
			ddlTracker.DDLProgress("sequence", seqName, "ok", nil)
		}
		slog.Info("sequence ddl applied", "sequence", seqName, "object_group", cfg.ObjectGroup)
	}

	return nil
}

func publishSequenceError(tracker ProgressTracker, ddlTracker DDLProgressTracker, hasDDLTracker bool, seqName string, err error) error {
	slog.Warn("failed to execute sequence DDL", "sequence", seqName, "error", err)
	if tracker != nil && tracker.EventBus() != nil {
		tracker.EventBus().Publish(bus.Event{
			Type:       bus.EventDDLProgress,
			Object:     "sequence",
			ObjectName: seqName,
			Status:     "error",
			Error:      err,
		})
	} else if hasDDLTracker {
		ddlTracker.DDLProgress("sequence", seqName, "error", err)
	}
	return fmt.Errorf("failed to execute sequence ddl %s: %w", seqName, err)
}

func classifyDDLGroup(objectType, ddl string) string {
	switch strings.ToLower(strings.TrimSpace(objectType)) {
	case "sequence":
		return config.ObjectGroupSequences
	case "table", "primary_key", "index", "constraint":
		return config.ObjectGroupTables
	}

	stmt := strings.ToUpper(strings.TrimSpace(ddl))
	switch {
	case strings.HasPrefix(stmt, "CREATE SEQUENCE"), strings.HasPrefix(stmt, "ALTER SEQUENCE"):
		return config.ObjectGroupSequences
	case strings.HasPrefix(stmt, "CREATE TABLE"),
		strings.HasPrefix(stmt, "ALTER TABLE"),
		strings.HasPrefix(stmt, "CREATE INDEX"),
		strings.HasPrefix(stmt, "CREATE UNIQUE INDEX"):
		return config.ObjectGroupTables
	default:
		slog.Warn("unable to classify ddl statement; defaulting to tables", "object_type", objectType)
		return config.ObjectGroupTables
	}
}
