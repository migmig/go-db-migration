package dialect

import (
	"fmt"
	"strings"
)

func (d *MySQLDialect) CreateTableDDL(tableName, schema string, cols []ColumnDef) string {
	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), fullTableName)
	}

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

func (d *MySQLDialect) CreateConstraintDDL(constraint ConstraintMetadata, schema string) string {
	table := strings.ToLower(constraint.TableName)
	if schema != "" {
		table = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(table))
	} else {
		table = d.QuoteIdentifier(table)
	}

	name := d.QuoteIdentifier(strings.ToLower(constraint.Name))

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s ", table, name))

	if constraint.Type == "R" {
		refTable := strings.ToLower(constraint.RefTableName)
		if schema != "" {
			refTable = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(refTable))
		} else {
			refTable = d.QuoteIdentifier(refTable)
		}

		localCols := make([]string, len(constraint.Columns))
		for i, c := range constraint.Columns {
			localCols[i] = d.QuoteIdentifier(strings.ToLower(c))
		}

		refCols := make([]string, len(constraint.RefColumns))
		for i, c := range constraint.RefColumns {
			refCols[i] = d.QuoteIdentifier(strings.ToLower(c))
		}

		sb.WriteString(fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
			strings.Join(localCols, ", "), refTable, strings.Join(refCols, ", ")))

		if constraint.DeleteRule != "" {
			sb.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.DeleteRule))
		}
	} else if constraint.Type == "C" {
		sb.WriteString(fmt.Sprintf("CHECK (%s)", constraint.SearchCondition))
	}
	sb.WriteString(";\n")
	return sb.String()
}

func (d *MySQLDialect) CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) {
	// MySQL doesn't support sequences like Oracle/Postgres. AUTO_INCREMENT is used on columns.
	return "", false
}

func (d *MySQLDialect) CreateIndexDDL(idx IndexMetadata, tableName, schema string) string {
	table := strings.ToLower(tableName)
	if schema != "" {
		table = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(table))
	} else {
		table = d.QuoteIdentifier(table)
	}

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
		return fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);\n", table, colList)
	}

	indexName := d.QuoteIdentifier(strings.ToLower(idx.Name))
	// MySQL 8+ supports IF NOT EXISTS for indexes in CREATE INDEX, but we can just use normal syntax.
	// We'll use IF NOT EXISTS since spec implies MySQL 8+.
	if idx.Uniqueness == "UNIQUE" {
		return fmt.Sprintf("CREATE UNIQUE INDEX %s ON %s (%s);\n", indexName, table, colList)
	}
	return fmt.Sprintf("CREATE INDEX %s ON %s (%s);\n", indexName, table, colList)
}
