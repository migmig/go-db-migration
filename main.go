package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/logger"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/security"
	"dbmigrator/internal/web"
)

func main() {
	if handled := handleUserCommand(os.Args[1:]); handled {
		return
	}

	cfg, err := config.ParseFlags()
	if err != nil {
		os.Exit(1)
	}

	if cfg.CompletionShell != "" {
		return
	}

	if cfg.WebMode {
		web.RunServerWithAuth("8080", cfg.AuthEnabled)
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
			pgPool, err := db.ConnectPostgres(cfg.TargetURL, cfg.DBMaxOpen, cfg.DBMaxIdle, cfg.DBMaxLife)
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

	report, err := migration.Run(oracleDB, targetDB, pool, dia, cfg, nil)
	if err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}
	_ = report // PrintSummary is called inside Run()
	slog.Info("migration completed successfully")
}

func handleUserCommand(args []string) bool {
	if len(args) == 0 || args[0] != "users" {
		return false
	}

	store, err := db.OpenUserStore(os.Getenv("DBM_AUTH_DB_PATH"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open user store: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	if len(args) < 2 {
		printUsersUsageAndExit()
	}

	sub := args[1]
	cmdArgs := args[2:]

	switch sub {
	case "list":
		users, err := store.ListUsers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to list users: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("ID\tUSERNAME\tADMIN\tCREATED_AT")
		for _, user := range users {
			fmt.Printf("%d\t%s\t%t\t%s\n", user.ID, user.Username, user.IsAdmin, user.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	case "add":
		if len(cmdArgs) < 2 || len(cmdArgs) > 3 {
			printUsersUsageAndExit()
		}
		username := strings.TrimSpace(cmdArgs[0])
		password := cmdArgs[1]
		isAdmin := len(cmdArgs) == 3 && cmdArgs[2] == "--admin"

		hash, err := security.HashPassword(password)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid password: %v\n", err)
			os.Exit(1)
		}
		if err := store.CreateUser(username, hash, isAdmin); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("user %q created\n", username)
	case "reset-password":
		if len(cmdArgs) != 2 {
			printUsersUsageAndExit()
		}
		username := strings.TrimSpace(cmdArgs[0])
		newPassword := cmdArgs[1]
		hash, err := security.HashPassword(newPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid password: %v\n", err)
			os.Exit(1)
		}
		if err := store.ResetPassword(username, hash); err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				fmt.Fprintf(os.Stderr, "user %q not found\n", username)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "failed to reset password: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("password reset for %q\n", username)
	case "delete":
		if len(cmdArgs) != 1 {
			printUsersUsageAndExit()
		}
		username := strings.TrimSpace(cmdArgs[0])
		if err := store.DeleteUser(username); err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				fmt.Fprintf(os.Stderr, "user %q not found\n", username)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "failed to delete user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("user %q deleted\n", username)
	default:
		printUsersUsageAndExit()
	}

	return true
}

func printUsersUsageAndExit() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  dbmigrator users list")
	fmt.Fprintln(os.Stderr, "  dbmigrator users add <username> <password> [--admin]")
	fmt.Fprintln(os.Stderr, "  dbmigrator users reset-password <username> <newPassword>")
	fmt.Fprintln(os.Stderr, "  dbmigrator users delete <username>")
	os.Exit(1)
}
