package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	go_ora "github.com/sijms/go-ora/v2"

	// Drivers for targets
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/microsoft/go-mssqldb"
)

// PGPool interface for testability
type PGPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Close()
}

func ConnectOracle(url, user, password string) (*sql.DB, error) {
	dsn := url
	if !strings.HasPrefix(dsn, "oracle://") {
		serverParts := strings.Split(dsn, "/")
		hostPort := serverParts[0]
		serviceName := ""
		if len(serverParts) > 1 {
			serviceName = serverParts[1]
		}

		host := hostPort
		port := 1521
		if strings.Contains(hostPort, ":") {
			parts := strings.Split(hostPort, ":")
			host = parts[0]
			fmt.Sscanf(parts[1], "%d", &port)
		}

		dsn = go_ora.BuildUrl(host, port, serviceName, user, password, nil)
	}

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectPostgres(url string, maxOpen int, maxIdle int, maxLife int) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, nil
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	if maxOpen > 0 {
		config.MaxConns = int32(maxOpen)
	}
	if maxIdle > 0 {
		config.MinConns = int32(maxIdle)
	}
	if maxLife > 0 {
		config.MaxConnLifetime = time.Duration(maxLife) * time.Second
	}

	return pgxpool.NewWithConfig(context.Background(), config)
}

func ConnectTargetDB(driverName, url string) (*sql.DB, error) {
	if url == "" {
		return nil, nil
	}
	db, err := sql.Open(driverName, url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func FetchTables(db *sql.DB, likeFilter string) ([]string, error) {
	query := `SELECT table_name FROM user_tables`
	var args []interface{}

	if likeFilter != "" {
		query += ` WHERE table_name LIKE :1`
		args = append(args, likeFilter)
	}
	query += ` ORDER BY table_name`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// FetchColumnTypes returns a map of lowercase column_name → data_type for the given PG table.
// data_type values are lowercase PostgreSQL type names from information_schema (e.g. "bigint", "numeric", "character varying").
func FetchColumnTypes(ctx context.Context, pool PGPool, schema, table string) (map[string]string, error) {
	query := `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema = $1
		  AND table_name   = lower($2)
		ORDER BY ordinal_position
	`
	rows, err := pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var colName, dataType string
		if err := rows.Scan(&colName, &dataType); err != nil {
			return nil, err
		}
		result[strings.ToLower(colName)] = strings.ToLower(dataType)
	}
	return result, rows.Err()
}

// SQLDBCountFn은 *sql.DB를 이용해 테이블 행 수를 조회하는 함수를 반환한다.
func SQLDBCountFn(d *sql.DB) func(ctx context.Context, tableName string) (int, error) {
	return func(ctx context.Context, tableName string) (int, error) {
		if d == nil {
			return 0, fmt.Errorf("db connection unavailable")
		}
		var count int
		err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
		return count, err
	}
}

// PGPoolCountFn은 PGPool을 이용해 테이블 행 수를 조회하는 함수를 반환한다.
func PGPoolCountFn(pool PGPool) func(ctx context.Context, tableName string) (int, error) {
	return func(ctx context.Context, tableName string) (int, error) {
		if pool == nil {
			return 0, fmt.Errorf("pg pool unavailable")
		}
		var count int
		err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
		return count, err
	}
}

func TableExists(ctx context.Context, pool PGPool, schema, table string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE  table_schema = $1
			AND    table_name   = lower($2)
		)
	`
	err := pool.QueryRow(ctx, query, schema, table).Scan(&exists)
	return exists, err
}
