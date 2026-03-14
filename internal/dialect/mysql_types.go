package dialect

import (
	"fmt"
	"strings"
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

func (d *MySQLDialect) DriverName() string {
	return "mysql"
}

func (d *MySQLDialect) NormalizeURL(url string) string {
	// target url typically user:pass@tcp(host:port)/db
	return url
}
