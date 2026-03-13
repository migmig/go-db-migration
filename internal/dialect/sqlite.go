package dialect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// SQLiteDialect implements Dialect for SQLite.
type SQLiteDialect struct{}

func (d *SQLiteDialect) Name() string {
	return "sqlite"
}

func (d *SQLiteDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func (d *SQLiteDialect) MapOracleType(oracleType string, precision, scale int) string {
	oracleType = strings.ToUpper(oracleType)

	switch {
	case strings.Contains(oracleType, "VARCHAR2") || strings.Contains(oracleType, "CHAR"):
		return "TEXT"
	case oracleType == "NUMBER":
		if precision > 0 {
			if scale > 0 {
				return "REAL"
			}
			return "INTEGER"
		}
		return "REAL"
	case oracleType == "DATE" || strings.Contains(oracleType, "TIMESTAMP"):
		return "TEXT"
	case oracleType == "CLOB":
		return "TEXT"
	case oracleType == "BLOB" || oracleType == "RAW":
		return "BLOB"
	case oracleType == "FLOAT":
		return "REAL"
	default:
		return "TEXT"
	}
}

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

func (d *SQLiteDialect) InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string {
	var stmts []string

	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))

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

func (d *SQLiteDialect) formatValue(val any) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		return fmt.Sprintf("x'%s'", hex.EncodeToString(v))
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05"))
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

func (d *SQLiteDialect) DriverName() string {
	return "sqlite3"
}

func (d *SQLiteDialect) NormalizeURL(url string) string {
	return url
}
