package config

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	OracleURL string
	User      string
	Password  string
	Tables    []string
	OutFile   string
	BatchSize int
	Schema    string
	PerTable  bool
	Parallel  bool
	// v2 flags
	PGURL   string
	Workers int
	WithDDL bool
	DryRun  bool
	LogJSON bool
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.OracleURL, "url", "", "Oracle Database URL (e.g., host:port/service_name)")
	flag.StringVar(&cfg.User, "user", "", "Database username")
	flag.StringVar(&cfg.Password, "password", "", "Database password")
	tablesFlag := flag.String("tables", "", "Comma-separated list of tables to migrate")
	flag.StringVar(&cfg.OutFile, "out", "migration.sql", "Output SQL file name")
	flag.IntVar(&cfg.BatchSize, "batch", 1000, "Number of rows per bulk insert")
	flag.StringVar(&cfg.Schema, "schema", "", "PostgreSQL schema name (optional)")
	flag.BoolVar(&cfg.PerTable, "per-table", false, "Output to separate files per table")
	flag.BoolVar(&cfg.Parallel, "parallel", false, "Process tables concurrently")
	// v2 flags
	flag.StringVar(&cfg.PGURL, "pg-url", "", "PostgreSQL Connection URL (e.g., postgres://user:pass@host:port/db)")
	flag.IntVar(&cfg.Workers, "workers", 4, "Number of concurrent workers for table processing")
	flag.BoolVar(&cfg.WithDDL, "with-ddl", false, "Generate CREATE TABLE DDLs")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "Verify connectivity and estimate data without actual migration")
	flag.BoolVar(&cfg.LogJSON, "log-json", false, "Enable structured JSON logging")

	flag.Parse()

	if cfg.OracleURL == "" || cfg.User == "" || cfg.Password == "" || *tablesFlag == "" {
		flag.Usage()
		return nil, fmt.Errorf("missing required flags")
	}

	t := strings.Split(*tablesFlag, ",")
	for i := range t {
		cfg.Tables = append(cfg.Tables, strings.TrimSpace(t[i]))
	}

	return cfg, nil
}
