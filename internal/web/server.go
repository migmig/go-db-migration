package web

import (
	"log"
	"net/http"

	"os"
	"path/filepath"
	"time"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/web/ws"
	"dbmigrator/internal/web/ziputil"

	"github.com/gin-gonic/gin"
)

var tracker = ws.NewWebSocketTracker()

func RunServer(port string) {
	r := gin.Default()

	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

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
}

func startMigration(c *gin.Context) {
	var req startMigrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	go func() {
		// Start migration process in background
		oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
		if err != nil {
			log.Printf("Failed to connect to Oracle: %v", err)
			return
		}
		defer oracleDB.Close()

		jobID := time.Now().Format("20060102150405")
		outDir := filepath.Join(os.TempDir(), "dbmigrator_"+jobID)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			log.Printf("Failed to create temp directory: %v", err)
			return
		}

		cfg := &config.Config{
			Tables:    req.Tables,
			Parallel:  true,
			Workers:   4,
			BatchSize: 1000,
			PerTable:  true,
			OutputDir: outDir,
		}

		err = migration.Run(oracleDB, nil, cfg, tracker)
		if err != nil {
			log.Printf("Migration failed: %v", err)
		} else {
			// Create ZIP
			zipFilePath := filepath.Join(os.TempDir(), "migration_"+jobID+".zip")
			if err := ziputil.ZipDirectory(outDir, zipFilePath); err != nil {
				log.Printf("Failed to create zip: %v", err)
			} else {
				tracker.AllDone("migration_" + jobID + ".zip")
			}
		}

		// Clean up the temporary SQL files folder (keep zip)
		os.RemoveAll(outDir)
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
