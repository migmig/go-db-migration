package migration

import (
	"database/sql"
	"fmt"
	"strings"

	"dbmigrator/internal/dialect"
)

func GetTableMetadata(db *sql.DB, tableName string) ([]dialect.ColumnDef, error) {
	query := `
		SELECT column_name, data_type, data_precision, data_scale, nullable
		FROM all_tab_columns
		WHERE table_name = :1
		ORDER BY column_id
	`
	rows, err := db.Query(query, strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []dialect.ColumnDef
	for rows.Next() {
		var c dialect.ColumnDef
		if err := rows.Scan(&c.Name, &c.Type, &c.Precision, &c.Scale, &c.Nullable); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func GenerateCreateTableDDL(tableName string, schema string, cols []dialect.ColumnDef, dia dialect.Dialect) string {
	return dia.CreateTableDDL(tableName, schema, cols)
}

// GetSequenceMetadata returns sequence metadata associated with the given table.
// It discovers sequences via DEFAULT column values and well-known naming patterns,
// then merges any explicitly listed names from extraNames.
func GetSequenceMetadata(db *sql.DB, tableName, owner string, extraNames []string) ([]dialect.SequenceMetadata, error) {
	tableUpper := strings.ToUpper(tableName)
	ownerUpper := strings.ToUpper(owner)

	// Collect candidate sequence names from DEFAULT column values (.NEXTVAL)
	defaultQuery := `
		SELECT DISTINCT REGEXP_REPLACE(data_default, '.*?([A-Z0-9_$#]+)\.NEXTVAL.*', '\1')
		FROM all_tab_columns
		WHERE owner = :1
		  AND table_name = :2
		  AND data_default IS NOT NULL
		  AND UPPER(data_default) LIKE '%.NEXTVAL%'
	`
	rows, err := db.Query(defaultQuery, ownerUpper, tableUpper)
	if err != nil {
		return nil, fmt.Errorf("sequence discovery query failed: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]bool)
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			continue
		}
		n = strings.ToUpper(strings.TrimSpace(n))
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}

	// Add well-known naming patterns
	patterns := []string{
		tableUpper + "_SEQ",
		tableUpper + "_ID_SEQ",
		"SEQ_" + tableUpper,
	}
	for _, p := range patterns {
		if !seen[p] {
			seen[p] = true
			names = append(names, p)
		}
	}

	// Merge explicitly specified names
	for _, n := range extraNames {
		n = strings.ToUpper(strings.TrimSpace(n))
		if n != "" && !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}

	if len(names) == 0 {
		return nil, nil
	}

	// Build IN clause placeholders
	placeholders := make([]string, len(names))
	args := make([]interface{}, 0, len(names)+1)
	args = append(args, ownerUpper)
	for i, n := range names {
		placeholders[i] = fmt.Sprintf(":%d", i+2)
		args = append(args, n)
	}

	seqQuery := fmt.Sprintf(`
		SELECT sequence_name, min_value, max_value, increment_by, cycle_flag, last_number
		FROM all_sequences
		WHERE sequence_owner = :1
		  AND sequence_name IN (%s)
		ORDER BY sequence_name
	`, strings.Join(placeholders, ","))

	seqRows, err := db.Query(seqQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("sequence metadata query failed: %w", err)
	}
	defer seqRows.Close()

	var seqs []dialect.SequenceMetadata
	for seqRows.Next() {
		var s dialect.SequenceMetadata
		if err := seqRows.Scan(&s.Name, &s.MinValue, &s.MaxValue, &s.IncrementBy, &s.CycleFlag, &s.LastNumber); err != nil {
			return nil, err
		}
		seqs = append(seqs, s)
	}
	return seqs, nil
}

// GenerateSequenceDDL converts an Oracle SequenceMetadata into a CREATE SEQUENCE statement.
// Returns an empty string and false if the dialect does not support sequences.
func GenerateSequenceDDL(seq dialect.SequenceMetadata, schema string, dia dialect.Dialect) (string, bool) {
	return dia.CreateSequenceDDL(seq, schema)
}

// GetIndexMetadata returns index metadata for the given table (excluding PK system indexes and LOB indexes).
func GetIndexMetadata(db *sql.DB, tableName, owner string) ([]dialect.IndexMetadata, error) {
	tableUpper := strings.ToUpper(tableName)
	ownerUpper := strings.ToUpper(owner)

	indexQuery := `
		SELECT i.index_name, i.uniqueness, i.index_type
		FROM all_indexes i
		WHERE i.table_owner = :1
		  AND i.table_name  = :2
		  AND i.index_type IN ('NORMAL', 'FUNCTION-BASED NORMAL')
		  AND NOT EXISTS (
		      SELECT 1 FROM all_lobs l
		      WHERE l.owner = i.table_owner
		        AND l.table_name = i.table_name
		        AND l.index_name = i.index_name
		  )
		ORDER BY i.index_name
	`
	rows, err := db.Query(indexQuery, ownerUpper, tableUpper)
	if err != nil {
		return nil, fmt.Errorf("index metadata query failed: %w", err)
	}
	defer rows.Close()

	var indexes []dialect.IndexMetadata
	for rows.Next() {
		var idx dialect.IndexMetadata
		if err := rows.Scan(&idx.Name, &idx.Uniqueness, &idx.IndexType); err != nil {
			return nil, err
		}
		idx.IsPK = strings.HasPrefix(idx.Name, "SYS_C")
		indexes = append(indexes, idx)
	}

	// Fetch columns for each index
	colQuery := `
		SELECT column_name, column_position, descend
		FROM all_ind_columns
		WHERE index_owner = :1
		  AND index_name  = :2
		ORDER BY column_position
	`
	for i := range indexes {
		colRows, err := db.Query(colQuery, ownerUpper, indexes[i].Name)
		if err != nil {
			return nil, fmt.Errorf("index column query failed for %s: %w", indexes[i].Name, err)
		}
		for colRows.Next() {
			var col dialect.IndexColumn
			if err := colRows.Scan(&col.Name, &col.Position, &col.Descend); err != nil {
				colRows.Close()
				return nil, err
			}
			indexes[i].Columns = append(indexes[i].Columns, col)
		}
		colRows.Close()
	}

	return indexes, nil
}

// GenerateIndexDDL converts an Oracle IndexMetadata into a DDL statement using the dialect.
func GenerateIndexDDL(idx dialect.IndexMetadata, tableName, schema string, dia dialect.Dialect) string {
	return dia.CreateIndexDDL(idx, tableName, schema)
}
