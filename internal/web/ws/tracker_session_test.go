package ws

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("expected session manager")
	}

	sessionID := sm.CreateSession()
	if sessionID == "" {
		t.Error("expected valid session ID")
	}

	tracker := sm.GetTracker(sessionID)
	if tracker == nil {
		t.Error("expected to find tracker for session")
	}

	nilTracker := sm.GetTracker("unknown")
	if nilTracker != nil {
		t.Error("expected nil for unknown session")
	}

	// Test cleanup loop
	// Give it a short time so the cleanup loop is started (it's a background goroutine)
	time.Sleep(50 * time.Millisecond)
}

func TestTracker_PartialSuccess(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.PartialSuccess("USERS", 1, 100)
}

func TestTracker_ValidationStart(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.ValidationStart("USERS")
}

func TestTracker_ValidationResult(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.ValidationResult("USERS", 100, 100, "success", "")
}

func TestTracker_DiscoverySummary(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.DiscoverySummary("tables", []string{"A"}, []string{"S"})
}

func TestSessionManager_HandleConnection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sm := NewSessionManager()
	sessionID := sm.CreateSession()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/ws?sessionId="+sessionID, nil)

	// This will fail because the request is not a websocket upgrade,
	// but it will cover the beginning of HandleConnection.
	sm.HandleConnection(c)
	if w.Code != http.StatusBadRequest {
		// Gorilla websocket returns 400 Bad Request if not a websocket request
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Missing sessionId
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request, _ = http.NewRequest("GET", "/ws", nil)
	sm.HandleConnection(c2)
	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w2.Code)
	}
}

func TestTracker_HandleConnection_MissingUpgrade(t *testing.T) {
	tr := NewWebSocketTracker()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/ws", nil)

	tr.HandleConnection(c)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTracker_Error(t *testing.T) {
    tr := NewWebSocketTracker()
    tr.Init("USERS", 200)
    tr.Error("USERS", errors.New("test error"))
}
