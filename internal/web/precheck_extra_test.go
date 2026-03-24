package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPrecheckHandler_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/precheck", precheckHandler(nil))

	t.Run("Invalid Policy", func(t *testing.T) {
		body := []byte(`{"policy":"invalid","tables":["T1"]}`)
		req, _ := http.NewRequest("POST", "/precheck", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Empty Tables", func(t *testing.T) {
		body := []byte(`{"policy":"strict","tables":[]}`)
		req, _ := http.NewRequest("POST", "/precheck", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("Oracle Connection Error", func(t *testing.T) {
		body := []byte(`{"policy":"strict","tables":["T1"],"oracleUrl":"invalid","username":"u","password":"p"}`)
		req, _ := http.NewRequest("POST", "/precheck", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}
