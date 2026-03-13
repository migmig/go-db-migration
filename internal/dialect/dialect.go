package dialect

import (
	"database/sql"
	"fmt"
)

// ColumnDef represents an Oracle column definition.
type ColumnDef struct {
	Name      string
	Type      string
	Precision sql.NullInt64
	Scale     sql.NullInt64
	Nullable  string
}

// SequenceMetadata holds Oracle sequence information.
type SequenceMetadata struct {
	Name        string
	MinValue    int64
	MaxValue    string
	IncrementBy int64
	CycleFlag   string
	LastNumber  int64
}

// IndexColumn represents a single column in an Oracle index.
type IndexColumn struct {
	Name     string
	Position int
	Descend  string
}

// IndexMetadata holds Oracle index information.
type IndexMetadata struct {
	Name       string
	Uniqueness string
	IndexType  string
	IsPK       bool
	Columns    []IndexColumn
}

// Dialect defines the interface for different target database dialects.
type Dialect interface {
	// Name returns the dialect name (e.g., "postgres", "mysql").
	Name() string

	// QuoteIdentifier quotes an identifier (e.g., table or column name) according to the dialect.
	QuoteIdentifier(name string) string

	// MapOracleType maps an Oracle data type to the target database type.
	MapOracleType(oracleType string, precision, scale int) string

	// CreateTableDDL generates the CREATE TABLE DDL.
	CreateTableDDL(tableName, schema string, cols []ColumnDef) string

	// CreateSequenceDDL generates the CREATE SEQUENCE DDL.
	// Returns a boolean indicating whether the target DB supports sequences.
	CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool)

	// CreateIndexDDL generates the CREATE INDEX DDL.
	CreateIndexDDL(idx IndexMetadata, tableName, schema string) string

	// InsertStatement generates batch INSERT statements.
	InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string

	// DriverName returns the Go SQL driver name (e.g., "pgx", "mysql", "sqlite3", "sqlserver").
	DriverName() string

	// NormalizeURL standardizes the connection URL for the target driver.
	NormalizeURL(url string) string
}

// GetDialect returns a dialect implementation based on the target DB name.
func GetDialect(name string) (Dialect, error) {
	switch name {
	case "postgres", "":
		return &PostgresDialect{}, nil
	case "mysql":
		return &MySQLDialect{}, nil
	case "mariadb":
		return &MariaDBDialect{}, nil
	case "sqlite":
		return &SQLiteDialect{}, nil
	case "mssql":
		return &MSSQLDialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported target database: %s", name)
	}
}
