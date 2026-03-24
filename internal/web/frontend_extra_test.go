package web

import (
	"os"
	"testing"
	"github.com/gin-gonic/gin"
)

func TestEmbeddedFrontendIsPlaceholder(t *testing.T) {
	// Use a real but empty FS
	dummy := os.DirFS(t.TempDir())
	_ = embeddedFrontendIsPlaceholder(dummy)
}

func TestRegisterFrontendRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Just check if it executes without panic
	_ = registerFrontendRoutes(r)
}

func TestRegisterFrontendRoutes_Nil(t *testing.T) {
	if registerFrontendRoutes(nil) {
		t.Error("expected false for nil engine")
	}
}
