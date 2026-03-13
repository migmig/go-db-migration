package dialect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// MySQLDialect implements Dialect for MySQL.
type MySQLDialect struct{}

func (d *MySQLDialect) Name() string {
	return "mysql"
}

func (d *MySQLDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("`%s`", name)
}

func (d *MySQLDialect) MapOracleType(oracleType string, precision, scale int) string {
	oracleType = strings.ToUpper(oracleType)

	switch {
	case strings.Contains(oracleType, "VARCHAR2"):
		if precision > 0 {
			if precision > 16383 {
				return "LONGTEXT"
			}
			return fmt.Sprintf("VARCHAR(%d)", precision)
		}
		return "VARCHAR(255)"

	case strings.Contains(oracleType, "CHAR"):
		if precision > 0 {
			return fmt.Sprintf("CHAR(%d)", precision)
		}
		return "CHAR(255)"
	case oracleType == "NUMBER":
		if precision > 0 {
			if scale > 0 {
				return fmt.Sprintf("DECIMAL(%d, %d)", precision, scale)
			}
			if precision <= 4 {
				return "SMALLINT"
			}
			if precision <= 9 {
				return "INT"
			}
			return "BIGINT"
		}
		return "DOUBLE"
	case oracleType == "DATE" || strings.Contains(oracleType, "TIMESTAMP"):
		return "DATETIME"
	case oracleType == "CLOB":
		return "LONGTEXT"
	case oracleType == "BLOB" || oracleType == "RAW":
		return "LONGBLOB"
	case oracleType == "FLOAT":
		return "DOUBLE"
	default:
		return "TEXT"
	}
}

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

func (d *MySQLDialect) InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string {
	var stmts []string

	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), fullTableName)
	}

	quotedCols := make([]string, len(cols))
	for i, c := range cols {
		quotedCols[i] = d.QuoteIdentifier(strings.ToLower(c))
	}
	colStr := strings.Join(quotedCols, ", ")

	baseStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES\n    ", fullTableName, colStr)

	var currentBatch []string
	for _, row := range rows {
		var rowVals []string
		for _, val := range row {
			rowVals = append(rowVals, d.formatValue(val))
		}
		currentBatch = append(currentBatch, fmt.Sprintf("(%s)", strings.Join(rowVals, ", ")))

		if len(currentBatch) >= batchSize {
			stmts = append(stmts, baseStmt+strings.Join(currentBatch, ",\n    ")+";\n")
			currentBatch = currentBatch[:0]
		}
	}

	if len(currentBatch) > 0 {
		stmts = append(stmts, baseStmt+strings.Join(currentBatch, ",\n    ")+";\n")
	}

	return stmts
}

func (d *MySQLDialect) formatValue(val any) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		return fmt.Sprintf("X'%s'", hex.EncodeToString(v))
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.999999"))
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	default:
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, "'", "''")
		escaped = strings.ReplaceAll(escaped, "\\", "\\\\")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func (d *MySQLDialect) DriverName() string {
	return "mysql"
}

func (d *MySQLDialect) NormalizeURL(url string) string {
	// target url typically user:pass@tcp(host:port)/db
	return url
}
