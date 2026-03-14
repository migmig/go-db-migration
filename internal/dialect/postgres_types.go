package dialect

import (
	"fmt"
	"strings"
)

// PostgresDialect implements Dialect for PostgreSQL.
type PostgresDialect struct{}

func (d *PostgresDialect) Name() string {
	return "postgres"
}

func (d *PostgresDialect) QuoteIdentifier(name string) string {
	return fmt.Sprintf("\"%s\"", name)
}

func (d *PostgresDialect) MapOracleType(oracleType string, precision, scale int) string {
	oracleType = strings.ToUpper(oracleType)

	switch {
	case strings.Contains(oracleType, "VARCHAR2") || strings.Contains(oracleType, "CHAR"):
		return "text"
	case oracleType == "NUMBER":
		if precision > 0 {
			if scale > 0 {
				return fmt.Sprintf("numeric(%d, %d)", precision, scale)
			}
			if precision <= 4 {
				return "smallint"
			}
			if precision <= 9 {
				return "integer"
			}
			return "bigint"
		}
		return "numeric"
	case oracleType == "DATE" || strings.Contains(oracleType, "TIMESTAMP"):
		return "timestamp"
	case oracleType == "CLOB":
		return "text"
	case oracleType == "BLOB" || oracleType == "RAW":
		return "bytea"
	case oracleType == "FLOAT":
		return "double precision"
	default:
		return "text"
	}
}

func (d *PostgresDialect) DriverName() string {
	return "pgx"
}

func (d *PostgresDialect) NormalizeURL(url string) string {
	return url
}
