package main

import (
	"database/sql"
	"log/slog"
	"os"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
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

	slog.Info("starting migration", "tables", cfg.Tables, "batch_size", cfg.BatchSize, "target_db", cfg.TargetDB)

	dia, err := dialect.GetDialect(cfg.TargetDB)
	if err != nil {
		slog.Error("failed to get dialect", "error", err)
		os.Exit(1)
	}

	oracleDB, err := db.ConnectOracle(cfg.OracleURL, cfg.User, cfg.Password)
	if err != nil {
		slog.Error("failed to connect to oracle", "error", err)
		os.Exit(1)
	}
	defer oracleDB.Close()

	var pool db.PGPool
	var targetDB *sql.DB

	if cfg.TargetURL != "" {
		if cfg.TargetDB == "postgres" {
			pgPool, err := db.ConnectPostgres(cfg.TargetURL)
			if err != nil {
				slog.Error("failed to connect to postgres", "error", err)
				os.Exit(1)
			}
			if pgPool != nil {
				pool = pgPool
				defer pgPool.Close()
				slog.Info("connected to postgres successfully")
			}
		} else {
			dbConn, err := db.ConnectTargetDB(dia.DriverName(), dia.NormalizeURL(cfg.TargetURL))
			if err != nil {
				slog.Error("failed to connect to target db", "error", err)
				os.Exit(1)
			}
			if dbConn != nil {
				targetDB = dbConn
				defer targetDB.Close()
				slog.Info("connected to target db successfully", "driver", dia.DriverName())
			}
		}
	}

	if err := migration.Run(oracleDB, targetDB, pool, dia, cfg, nil); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migration completed successfully")
}
