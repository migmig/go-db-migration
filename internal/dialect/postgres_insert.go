package dialect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func (d *PostgresDialect) InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string {
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

func (d *PostgresDialect) formatValue(val any) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		// BLOB/RAW to hex
		return fmt.Sprintf("'\\x%s'", hex.EncodeToString(v))
	case string:
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.999999999"))
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
		return fmt.Sprintf("'%s'", escaped)
	}
}
