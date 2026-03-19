package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/logger"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/security"
	"dbmigrator/internal/web"
)

var userCommandExit = os.Exit

func main() {
	if handled := handleUserCommand(os.Args[1:]); handled {
		return
	}

	cfg, err := config.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	if cfg.CompletionShell != "" {
		return
	}

	if cfg.WebMode {
		web.RunServerWithAuth("8080", cfg.AuthEnabled)
		return
	}

	logger.Setup(cfg.LogJSON)

	slog.Info("starting migration", "tables", cfg.Tables, "batch_size", cfg.BatchSize, "target_db", cfg.TargetDB)

	if cfg.TargetDB != "" && cfg.TargetDB != "postgres" {
		slog.Error("v22 이후 타겟 DB는 PostgreSQL만 지원합니다", "targetDb", cfg.TargetDB)
		os.Exit(1)
	}

	dia, err := dialect.GetDialect("postgres")
	if err != nil {
		slog.Error("failed to get dialect", "error", err)
		os.Exit(1)
	}

	oracleDB, err := db.ConnectOracle(cfg.OracleURL, cfg.User, cfg.Password)
	if err != nil {
		slog.Error("failed to connect to oracle", "error", err)
		os.Exit(1)
	}
	defer oracleDB.Close()

	var pool db.PGPool
	var targetDB *sql.DB

	if cfg.TargetURL != "" {
		pgPool, err := db.ConnectPostgres(cfg.TargetURL, cfg.DBMaxOpen, cfg.DBMaxIdle, cfg.DBMaxLife)
		if err != nil {
			slog.Error("failed to connect to postgres", "error", err)
			os.Exit(1)
		}
		if pgPool != nil {
			pool = pgPool
			defer pgPool.Close()
			slog.Info("connected to postgres successfully")
		}
	}

	// v19: pre-check 행 수 비교
	if cfg.PrecheckRowCount {
		precheckCfg := migration.PrecheckEngineConfig{
			Policy: migration.PrecheckPolicy(cfg.PrecheckPolicy),
		}
		slog.Info("running precheck row count", "tables", len(cfg.Tables), "policy", cfg.PrecheckPolicy)
		sourceCountFn := db.SQLDBCountFn(oracleDB, nil)
		var targetCountFn migration.RowCountFn
		if pool != nil {
			targetCountFn = db.PGPoolCountFn(pool, dia.QuoteIdentifier)
		}
		precheckResults, precheckSummary := migration.RunPrecheckRowCount(nil, cfg.Tables, sourceCountFn, targetCountFn, precheckCfg)
		slog.Info("precheck complete",
			"total", precheckSummary.TotalTables,
			"transfer_required", precheckSummary.TransferRequiredCount,
			"skip_candidate", precheckSummary.SkipCandidateCount,
			"count_check_failed", precheckSummary.CountCheckFailedCount,
		)
		filtered := migration.FilterPrecheckResults(precheckResults, cfg.PrecheckFilter)
		for _, r := range filtered {
			slog.Info("precheck item",
				"table_name", r.TableName,
				"decision", r.Decision,
				"source_row_count", r.SourceRowCount,
				"target_row_count", r.TargetRowCount,
				"diff", r.Diff,
				"reason", r.Reason,
			)
		}
		if cfg.DryRun {
			slog.Info("dry-run mode: skipping migration after precheck")
			return
		}
		// strict 정책에서 count_check_failed가 있으면 중단
		plan, planErr := migration.ApplyPrecheckPolicy(precheckResults, migration.PrecheckPolicy(cfg.PrecheckPolicy))
		if planErr != nil {
			slog.Error("precheck policy error", "error", planErr)
			os.Exit(1)
		}
		if plan.Blocked {
			slog.Error("precheck blocked migration", "reason", plan.BlockReason)
			os.Exit(1)
		}
		// skip_equal_rows 정책이면 transfer_tables만 마이그레이션
		if migration.PrecheckPolicy(cfg.PrecheckPolicy) == migration.PolicySkipEqualRows && len(plan.TransferTables) > 0 {
			cfg.Tables = plan.TransferTables
			slog.Info("precheck: reduced table list", "tables", cfg.Tables)
		}
	}

	report, err := migration.Run(oracleDB, targetDB, pool, dia, cfg, nil)
	if err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	_ = report // PrintSummary is called inside Run()
	slog.Info("migration completed successfully")
}

func handleUserCommand(args []string) bool {
	if len(args) == 0 || args[0] != "users" {
		return false
	}

	store, err := db.OpenUserStore(os.Getenv("DBM_AUTH_DB_PATH"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open user store: %v\n", err)
		userCommandExit(1)
	}
	defer store.Close()

	exitCode := executeUserCommand(store, args[1:], os.Stdout, os.Stderr)
	if exitCode != 0 {
		userCommandExit(exitCode)
	}

	return true
}

func executeUserCommand(store *db.UserStore, args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		printUsersUsage(stderr)
		return 1
	}

	sub := args[0]
	cmdArgs := args[1:]

	switch sub {
	case "list":
		users, err := store.ListUsers()
		if err != nil {
			fmt.Fprintf(stderr, "failed to list users: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "ID\tUSERNAME\tADMIN\tCREATED_AT")
		for _, user := range users {
			fmt.Fprintf(stdout, "%d\t%s\t%t\t%s\n", user.ID, user.Username, user.IsAdmin, user.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		return 0
	case "add":
		if len(cmdArgs) < 2 || len(cmdArgs) > 3 {
			printUsersUsage(stderr)
			return 1
		}
		username := strings.TrimSpace(cmdArgs[0])
		password := cmdArgs[1]
		isAdmin := len(cmdArgs) == 3 && cmdArgs[2] == "--admin"

		hash, err := security.HashPassword(password)
		if err != nil {
			fmt.Fprintf(stderr, "invalid password: %v\n", err)
			return 1
		}
		if err := store.CreateUser(username, hash, isAdmin); err != nil {
			fmt.Fprintf(stderr, "failed to add user: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "user %q created\n", username)
		return 0
	case "reset-password":
		if len(cmdArgs) != 2 {
			printUsersUsage(stderr)
			return 1
		}
		username := strings.TrimSpace(cmdArgs[0])
		newPassword := cmdArgs[1]
		hash, err := security.HashPassword(newPassword)
		if err != nil {
			fmt.Fprintf(stderr, "invalid password: %v\n", err)
			return 1
		}
		if err := store.ResetPassword(username, hash); err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				fmt.Fprintf(stderr, "user %q not found\n", username)
				return 1
			}
			fmt.Fprintf(stderr, "failed to reset password: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "password reset for %q\n", username)
		return 0
	case "delete":
		if len(cmdArgs) != 1 {
			printUsersUsage(stderr)
			return 1
		}
		username := strings.TrimSpace(cmdArgs[0])
		if err := store.DeleteUser(username); err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				fmt.Fprintf(stderr, "user %q not found\n", username)
				return 1
			}
			fmt.Fprintf(stderr, "failed to delete user: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "user %q deleted\n", username)
		return 0
	default:
		printUsersUsage(stderr)
		return 1
	}
}

func printUsersUsage(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  dbmigrator users list")
	fmt.Fprintln(w, "  dbmigrator users add <username> <password> [--admin]")
	fmt.Fprintln(w, "  dbmigrator users reset-password <username> <newPassword>")
	fmt.Fprintln(w, "  dbmigrator users delete <username>")
}

func runUserCommandForTest(store *db.UserStore, args []string) (int, string, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := executeUserCommand(store, args, &stdout, &stderr)
	return exitCode, stdout.String(), stderr.String()
}
