package dialect

import (
	"fmt"
	"strings"
)

func (d *SQLiteDialect) CreateTableDDL(tableName, schema string, cols []ColumnDef) string {
	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", fullTableName))

	for i, col := range cols {
		prec := 0
		if col.Precision.Valid {
			prec = int(col.Precision.Int64)
		}
		s := 0
		if col.Scale.Valid {
			s = int(col.Scale.Int64)
		}

		myType := d.MapOracleType(col.Type, prec, s)
		sb.WriteString(fmt.Sprintf("    %s %s", d.QuoteIdentifier(strings.ToLower(col.Name)), myType))

		if col.DefaultValue.Valid && col.DefaultValue.String != "" {
			sb.WriteString(fmt.Sprintf(" DEFAULT %s", col.DefaultValue.String))
		}

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

func (d *SQLiteDialect) CreateConstraintDDL(constraint ConstraintMetadata, schema string) string {
	// SQLite does not support ALTER TABLE ADD CONSTRAINT.
	// Constraints must be defined in CREATE TABLE or by recreating the table.
	// For this tool, we will just return an empty string or a comment, as we cannot add it post-creation.
	return fmt.Sprintf("-- SQLite does not support ALTER TABLE ADD CONSTRAINT for %s\n", constraint.Name)
}

func (d *SQLiteDialect) CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) {
	// SQLite uses AUTOINCREMENT keywords, no sequences
	return "", false
}

func (d *SQLiteDialect) CreateIndexDDL(idx IndexMetadata, tableName, schema string) string {
	table := d.QuoteIdentifier(strings.ToLower(tableName))

	colExprs := make([]string, len(idx.Columns))
	for i, col := range idx.Columns {
		expr := d.QuoteIdentifier(strings.ToLower(col.Name))
		if strings.ToUpper(col.Descend) == "DESC" {
			expr += " DESC"
		}
		colExprs[i] = expr
	}
	colList := strings.Join(colExprs, ", ")

	if idx.IsPK {
		// SQLite generally handles PK in CREATE TABLE, but we can't easily alter here. We skip or use CREATE UNIQUE INDEX
		return ""
	}

	indexName := d.QuoteIdentifier(strings.ToLower(idx.Name))

	if idx.Uniqueness == "UNIQUE" {
		return fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s);\n", indexName, table, colList)
	}
	return fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s);\n", indexName, table, colList)
}
