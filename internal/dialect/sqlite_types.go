package dialect

import (
	"fmt"
	"strings"
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

func (d *SQLiteDialect) DriverName() string {
	return "sqlite3"
}

func (d *SQLiteDialect) NormalizeURL(url string) string {
	return url
}
