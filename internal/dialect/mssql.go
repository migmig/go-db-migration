package dialect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// MSSQLDialect implements Dialect for MSSQL.
type MSSQLDialect struct{}

func (d *MSSQLDialect) Name() string {
	return "mssql"
}

func (d *MSSQLDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("[%s]", name)
}

func (d *MSSQLDialect) MapOracleType(oracleType string, precision, scale int) string {
	oracleType = strings.ToUpper(oracleType)

	switch {
	case strings.Contains(oracleType, "VARCHAR2"):
		return "NVARCHAR(MAX)" // Use MAX for broad compat
	case strings.Contains(oracleType, "CHAR"):
		return "NCHAR(255)"
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
		return "FLOAT"
	case oracleType == "DATE" || strings.Contains(oracleType, "TIMESTAMP"):
		return "DATETIME2"
	case oracleType == "CLOB":
		return "NVARCHAR(MAX)"
	case oracleType == "BLOB" || oracleType == "RAW":
		return "VARBINARY(MAX)"
	case oracleType == "FLOAT":
		return "FLOAT"
	default:
		return "NVARCHAR(MAX)"
	}
}

func (d *MSSQLDialect) CreateTableDDL(tableName, schema string, cols []ColumnDef) string {
	fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), fullTableName)
	}

	var sb strings.Builder

	// Check existence logic for MSSQL
	bareTableName := strings.ToLower(tableName)
	sb.WriteString(fmt.Sprintf("IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '%s')\n", bareTableName))
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", fullTableName))

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

func (d *MSSQLDialect) CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) {
	name := strings.ToLower(seq.Name)
	if schema != "" {
		name = fmt.Sprintf("%s.%s", d.QuoteIdentifier(strings.ToLower(schema)), d.QuoteIdentifier(name))
	} else {
		name = d.QuoteIdentifier(name)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE SEQUENCE %s\n", name))
	sb.WriteString(fmt.Sprintf("    START WITH %d\n", seq.LastNumber))
	sb.WriteString(fmt.Sprintf("    INCREMENT BY %d\n", seq.IncrementBy))
	sb.WriteString(fmt.Sprintf("    MINVALUE %d\n", seq.MinValue))

	if strings.TrimSpace(seq.MaxValue) != oracleDefaultMaxValue {
		sb.WriteString(fmt.Sprintf("    MAXVALUE %s\n", strings.TrimSpace(seq.MaxValue)))
	}

	if seq.CycleFlag == "Y" {
		sb.WriteString("    CYCLE\n")
	} else {
		sb.WriteString("    NO CYCLE\n")
	}

	sb.WriteString(";\n")
	return sb.String(), true
}

func (d *MSSQLDialect) CreateIndexDDL(idx IndexMetadata, tableName, schema string) string {
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

	indexName := strings.ToLower(idx.Name)
	quotedIndexName := d.QuoteIdentifier(indexName)

	uniqueStr := ""
	if idx.Uniqueness == "UNIQUE" {
		uniqueStr = "UNIQUE "
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = '%s')\n", indexName))
	sb.WriteString(fmt.Sprintf("    CREATE %sINDEX %s ON %s (%s);\n", uniqueStr, quotedIndexName, table, colList))

	return sb.String()
}

func (d *MSSQLDialect) InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string {
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

	// MSSQL batch limit is 1000. Force batchSize to max 1000
	effectiveBatchSize := batchSize
	if effectiveBatchSize > 1000 {
		effectiveBatchSize = 1000
	}

	var currentBatch []string
	for _, row := range rows {
		var rowVals []string
		for _, val := range row {
			rowVals = append(rowVals, d.formatValue(val))
		}
		currentBatch = append(currentBatch, fmt.Sprintf("(%s)", strings.Join(rowVals, ", ")))

		if len(currentBatch) >= effectiveBatchSize {
			stmts = append(stmts, baseStmt+strings.Join(currentBatch, ",\n    ")+";\n")
			currentBatch = currentBatch[:0]
		}
	}

	if len(currentBatch) > 0 {
		stmts = append(stmts, baseStmt+strings.Join(currentBatch, ",\n    ")+";\n")
	}

	return stmts
}

func (d *MSSQLDialect) formatValue(val any) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		return fmt.Sprintf("0x%s", hex.EncodeToString(v))
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("N'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.9999999"))
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
		return fmt.Sprintf("N'%s'", escaped)
	}
}

func (d *MSSQLDialect) DriverName() string {
	return "sqlserver"
}

func (d *MSSQLDialect) NormalizeURL(url string) string {
	return url
}
