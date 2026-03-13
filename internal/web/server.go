package web

import (
	"database/sql"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"os"
	"path/filepath"
	"time"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/logger"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/web/ws"
	"dbmigrator/internal/web/ziputil"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templateFS embed.FS

var tracker = ws.NewWebSocketTracker()

func RunServer(port string) {
	r := gin.Default()

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Oracle to PostgreSQL Migrator",
		})
	})

	api := r.Group("/api")
	{
		api.POST("/tables", getTables)
		api.POST("/migrate", startMigration)
		api.GET("/progress", tracker.HandleConnection)
		api.GET("/download/:id", downloadZip)
	}

	log.Printf("Starting web server on port %s...", port)
	if err := r.Run("localhost:" + port); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

type getTablesRequest struct {
	OracleURL string `json:"oracleUrl" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Like      string `json:"like"`
}

func getTables(c *gin.Context) {
	var req getTablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Oracle DB: " + err.Error()})
		return
	}
	defer oracleDB.Close()

	tables, err := db.FetchTables(oracleDB, req.Like)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

type startMigrationRequest struct {
	OracleURL string   `json:"oracleUrl" binding:"required"`
	Username  string   `json:"username" binding:"required"`
	Password  string   `json:"password" binding:"required"`
	Tables    []string `json:"tables" binding:"required"`
	Direct    bool     `json:"direct"`
	PGURL     string   `json:"pgUrl"`
	WithDDL   bool     `json:"withDdl"`
	BatchSize int      `json:"batchSize"`
	Workers   int      `json:"workers"`
	// v4 추가 필드
	OutFile  string `json:"outFile"`
	PerTable bool   `json:"perTable"`
	Schema   string `json:"schema"`
	DryRun   bool   `json:"dryRun"`
	LogJSON  bool   `json:"logJson"`
	// v5 추가 필드
	WithSequences bool   `json:"withSequences"`
	WithIndexes   bool   `json:"withIndexes"`
	OracleOwner   string `json:"oracleOwner"`
	// v6 추가 필드
	TargetDB  string `json:"targetDb"`
	TargetURL string `json:"targetUrl"`
}

var schemaPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateMigrationRequest(req *startMigrationRequest) error {
	if strings.ContainsAny(req.OutFile, `/\`) {
		return fmt.Errorf("outFile must not contain path separators")
	}
	if req.Schema != "" && !schemaPattern.MatchString(req.Schema) {
		return fmt.Errorf("schema name contains invalid characters")
	}
	if req.BatchSize < 0 {
		return fmt.Errorf("batchSize must be non-negative")
	}
	if req.Workers < 0 {
		return fmt.Errorf("workers must be non-negative")
	}
	return nil
}

func startMigration(c *gin.Context) {
	var req startMigrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	if err := validateMigrationRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go func() {
		if req.LogJSON {
			logger.SetJSONMode(true)
			defer logger.SetJSONMode(false)
		}

		// Start migration process in background
		oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
		if err != nil {
			log.Printf("Failed to connect to Oracle: %v", err)
			tracker.AllDone("")
			return
		}
		defer oracleDB.Close()

		targetDBName := req.TargetDB
		if targetDBName == "" {
			targetDBName = "postgres"
		}

		dia, err := dialect.GetDialect(targetDBName)
		if err != nil {
			log.Printf("Failed to get dialect: %v", err)
			tracker.AllDone("")
			return
		}

		var pgPool db.PGPool
		var targetDB *sql.DB

		targetURL := req.TargetURL
		if targetURL == "" {
			targetURL = req.PGURL // backward compat
		}

		if req.Direct && targetURL != "" {
			if targetDBName == "postgres" {
				pgPool, err = db.ConnectPostgres(targetURL)
				if err != nil {
					log.Printf("Failed to connect to Postgres: %v", err)
					tracker.AllDone("")
					return
				}
				defer pgPool.Close()
			} else {
				targetDB, err = db.ConnectTargetDB(dia.DriverName(), dia.NormalizeURL(targetURL))
				if err != nil {
					log.Printf("Failed to connect to Target DB: %v", err)
					tracker.AllDone("")
					return
				}
				defer targetDB.Close()
			}
		}

		workers := req.Workers
		if workers <= 0 {
			workers = 4
		}
		batchSize := req.BatchSize
		if batchSize <= 0 {
			batchSize = 1000
		}
		outFile := req.OutFile
		if outFile == "" {
			outFile = "migration.sql"
		}

		jobID := time.Now().Format("20060102150405")
		outDir := filepath.Join(os.TempDir(), "dbmigrator_"+jobID)
		if !req.Direct && !req.DryRun {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				log.Printf("Failed to create temp directory: %v", err)
				return
			}
		}

		cfg := &config.Config{
			Tables:        req.Tables,
			Parallel:      true,
			Workers:       workers,
			BatchSize:     batchSize,
			PerTable:      req.PerTable,
			OutFile:       outFile,
			Schema:        req.Schema,
			DryRun:        req.DryRun,
			OutputDir:     outDir,
			TargetDB:      targetDBName,
			TargetURL:     targetURL,
			WithDDL:       req.WithDDL,
			WithSequences: req.WithSequences,
			WithIndexes:   req.WithIndexes,
			OracleOwner:   req.OracleOwner,
		}

		err = migration.Run(oracleDB, targetDB, pgPool, dia, cfg, tracker)
		if err != nil {
			log.Printf("Migration failed: %v", err)
			tracker.AllDone("")
		} else if req.DryRun {
			tracker.AllDone("")
		} else if !req.Direct {
			// Create ZIP
			zipFilePath := filepath.Join(os.TempDir(), "migration_"+jobID+".zip")
			if err := ziputil.ZipDirectory(outDir, zipFilePath); err != nil {
				log.Printf("Failed to create zip: %v", err)
				tracker.AllDone("")
			} else {
				tracker.AllDone("migration_" + jobID + ".zip")
			}
		} else {
			tracker.AllDone("")
		}

		// Clean up the temporary SQL files folder (keep zip)
		if !req.Direct && !req.DryRun {
			os.RemoveAll(outDir)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Migration started"})
}

func downloadZip(c *gin.Context) {
	id := filepath.Base(c.Param("id")) // prevent path traversal
	if id == "" || id == "." || id == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing zip file ID"})
		return
	}

	zipPath := filepath.Join(os.TempDir(), id)
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+id)
	c.Header("Content-Type", "application/zip")
	c.File(zipPath)

	// Clean up zip after download
	// Wait a moment before deleting to ensure transfer completes
	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(zipPath)
	}()
}
