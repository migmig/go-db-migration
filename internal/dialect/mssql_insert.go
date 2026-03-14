package dialect

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

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
