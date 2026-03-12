package main

import (
	"log/slog"
	"os"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/logger"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/web"
)

func main() {
	cfg, err := config.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	if cfg.WebMode {
		web.RunServer("8080")
		return
	}

	logger.Setup(cfg.LogJSON)

	slog.Info("starting migration", "tables", cfg.Tables, "batch_size", cfg.BatchSize)

	oracleDB, err := db.ConnectOracle(cfg.OracleURL, cfg.User, cfg.Password)
	if err != nil {
		slog.Error("failed to connect to oracle", "error", err)
		os.Exit(1)
	}
	defer oracleDB.Close()

	pgPool, err := db.ConnectPostgres(cfg.PGURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	if pgPool != nil {
		defer pgPool.Close()
		slog.Info("connected to postgres successfully")
	}

	if err := migration.Run(oracleDB, pgPool, cfg, nil); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migration completed successfully")
}
