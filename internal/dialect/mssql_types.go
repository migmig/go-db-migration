package dialect

import (
	"fmt"
	"strings"
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
		if precision > 0 {
			if precision <= 4000 {
				return fmt.Sprintf("NVARCHAR(%d)", precision)
			}
			return "NVARCHAR(MAX)"
		}
		return "NVARCHAR(MAX)"

	case strings.Contains(oracleType, "CHAR"):
		if precision > 0 {
			if precision > 4000 {
				return "NCHAR(4000)"
			}
			return fmt.Sprintf("NCHAR(%d)", precision)
		}
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
		return "NUMERIC"
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

func (d *MSSQLDialect) DriverName() string {
	return "sqlserver"
}

func (d *MSSQLDialect) NormalizeURL(url string) string {
	return url
}
