package migration

import (
	"database/sql"
	"fmt"
	"strings"
)

// Oracle 기본 MAXVALUE (28자리 9)
const oracleDefaultMaxValue = "9999999999999999999999999999"

// SequenceMetadata holds Oracle sequence information.
type SequenceMetadata struct {
	Name        string
	MinValue    int64
	MaxValue    string
	IncrementBy int64
	CycleFlag   string
	LastNumber  int64
}

// IndexColumn represents a single column in an Oracle index.
type IndexColumn struct {
	Name     string
	Position int
	Descend  string
}

// IndexMetadata holds Oracle index information.
type IndexMetadata struct {
	Name       string
	Uniqueness string
	IndexType  string
	IsPK       bool
	Columns    []IndexColumn
}

type ColumnMetadata struct {
	Name      string
	Type      string
	Precision sql.NullInt64
	Scale     sql.NullInt64
	Nullable  string
}

func GetTableMetadata(db *sql.DB, tableName string) ([]ColumnMetadata, error) {
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

	var cols []ColumnMetadata
	for rows.Next() {
		var c ColumnMetadata
		if err := rows.Scan(&c.Name, &c.Type, &c.Precision, &c.Scale, &c.Nullable); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func GenerateCreateTableDDL(tableName string, schema string, cols []ColumnMetadata) string {
	fullTableName := tableName
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", schema, tableName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", fullTableName))

	for i, col := range cols {
		pgType := MapOracleToPostgres(col)
		sb.WriteString(fmt.Sprintf("    %s %s", strings.ToLower(col.Name), pgType))
		if col.Nullable == "N" {
			sb.WriteString(" NOT NULL")
		}
		if i < len(cols)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(");\n")
	return sb.String()
}

func MapOracleToPostgres(col ColumnMetadata) string {
	oracleType := strings.ToUpper(col.Type)

	switch {
	case strings.Contains(oracleType, "VARCHAR2") || strings.Contains(oracleType, "CHAR"):
		return "text"
	case oracleType == "NUMBER":
		if col.Precision.Valid {
			if col.Scale.Valid && col.Scale.Int64 > 0 {
				return fmt.Sprintf("numeric(%d, %d)", col.Precision.Int64, col.Scale.Int64)
			}
			if col.Precision.Int64 <= 9 {
				return "integer"
			}
			return "bigint"
		}
		return "numeric"
	case oracleType == "DATE" || strings.Contains(oracleType, "TIMESTAMP"):
		return "timestamp"
	case oracleType == "CLOB":
		return "text"
	case oracleType == "BLOB" || oracleType == "RAW":
		return "bytea"
	case oracleType == "FLOAT":
		return "double precision"
	default:
		return "text"
	}
}

// GetSequenceMetadata returns sequence metadata associated with the given table.
// It discovers sequences via DEFAULT column values and well-known naming patterns,
// then merges any explicitly listed names from extraNames.
func GetSequenceMetadata(db *sql.DB, tableName, owner string, extraNames []string) ([]SequenceMetadata, error) {
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

	var seqs []SequenceMetadata
	for seqRows.Next() {
		var s SequenceMetadata
		if err := seqRows.Scan(&s.Name, &s.MinValue, &s.MaxValue, &s.IncrementBy, &s.CycleFlag, &s.LastNumber); err != nil {
			return nil, err
		}
		seqs = append(seqs, s)
	}
	return seqs, nil
}

// GenerateSequenceDDL converts an Oracle SequenceMetadata into a PostgreSQL CREATE SEQUENCE statement.
func GenerateSequenceDDL(seq SequenceMetadata, schema string) string {
	name := strings.ToLower(seq.Name)
	if schema != "" {
		name = fmt.Sprintf("%s.%s", schema, name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE SEQUENCE IF NOT EXISTS %s\n", name))
	sb.WriteString(fmt.Sprintf("    START WITH %d\n", seq.LastNumber))
	sb.WriteString(fmt.Sprintf("    INCREMENT BY %d\n", seq.IncrementBy))
	sb.WriteString(fmt.Sprintf("    MINVALUE %d\n", seq.MinValue))

	// Omit MAXVALUE when it equals Oracle's default (28 nines)
	if strings.TrimSpace(seq.MaxValue) != oracleDefaultMaxValue {
		sb.WriteString(fmt.Sprintf("    MAXVALUE %s\n", strings.TrimSpace(seq.MaxValue)))
	}

	if seq.CycleFlag == "Y" {
		sb.WriteString("    CYCLE\n")
	} else {
		sb.WriteString("    NO CYCLE\n")
	}

	sb.WriteString(";\n")
	return sb.String()
}

// GetIndexMetadata returns index metadata for the given table (excluding PK system indexes and LOB indexes).
func GetIndexMetadata(db *sql.DB, tableName, owner string) ([]IndexMetadata, error) {
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

	var indexes []IndexMetadata
	for rows.Next() {
		var idx IndexMetadata
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
			var col IndexColumn
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

// GenerateIndexDDL converts an Oracle IndexMetadata into a PostgreSQL DDL statement.
func GenerateIndexDDL(idx IndexMetadata, tableName, schema string) string {
	table := strings.ToLower(tableName)
	if schema != "" {
		table = fmt.Sprintf("%s.%s", schema, table)
	}

	colExprs := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		expr := strings.ToLower(col.Name)
		if strings.ToUpper(col.Descend) == "DESC" {
			expr += " DESC"
		}
		colExprs[i] = expr
	}
	colList := strings.Join(colExprs, ", ")

	if idx.IsPK {
		return fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);\n", table, colList)
	}

	indexName := strings.ToLower(idx.Name)
	if idx.Uniqueness == "UNIQUE" {
		return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s);\n", indexName, table, colList)
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s);\n", indexName, table, colList)
}
