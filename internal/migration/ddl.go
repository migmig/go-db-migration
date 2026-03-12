package migration

import (
	"database/sql"
	"fmt"
	"strings"
)

type ColumnMetadata struct {
	Name      string
	Type      string
	Precision sql.NullInt64
	Scale     sql.NullInt64
	Nullable  string
}

func GetTableMetadata(db *sql.DB, tableName string) ([]ColumnMetadata, error) {
	query := `
		SELECT column_name, data_type, data_precision, data_scale, nullable
		FROM all_tab_columns
		WHERE table_name = :1
		ORDER BY column_id
	`
	rows, err := db.Query(query, strings.ToUpper(tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []ColumnMetadata
	for rows.Next() {
		var c ColumnMetadata
		if err := rows.Scan(&c.Name, &c.Type, &c.Precision, &c.Scale, &c.Nullable); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}

func GenerateCreateTableDDL(tableName string, schema string, cols []ColumnMetadata) string {
	fullTableName := tableName
	if schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", schema, tableName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", fullTableName))

	for i, col := range cols {
		pgType := MapOracleToPostgres(col)
		sb.WriteString(fmt.Sprintf("    %s %s", strings.ToLower(col.Name), pgType))
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

func MapOracleToPostgres(col ColumnMetadata) string {
	oracleType := strings.ToUpper(col.Type)

	switch {
	case strings.Contains(oracleType, "VARCHAR2") || strings.Contains(oracleType, "CHAR"):
		return "text"
	case oracleType == "NUMBER":
		if col.Precision.Valid {
			if col.Scale.Valid && col.Scale.Int64 > 0 {
				return fmt.Sprintf("numeric(%d, %d)", col.Precision.Int64, col.Scale.Int64)
			}
			if col.Precision.Int64 <= 9 {
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
