package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	go_ora "github.com/sijms/go-ora/v2"
)

// PGPool interface for testability
type PGPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
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

func ConnectPostgres(url string) (*pgxpool.Pool, error) {
	if url == "" {
		return nil, nil
	}
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	return pgxpool.NewWithConfig(context.Background(), config)
}

func TableExists(ctx context.Context, pool PGPool, schema, table string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE  table_schema = $1
			AND    table_name   = $2
		)
	`
	err := pool.QueryRow(ctx, query, schema, table).Scan(&exists)
	return exists, err
}
