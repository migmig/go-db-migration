package ws

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// ── constructor ───────────────────────────────────────────────────────────────

func TestNewWebSocketTracker(t *testing.T) {
	tr := NewWebSocketTracker()
	if tr == nil {
		t.Fatal("expected non-nil tracker")
	}
	if tr.clients == nil {
		t.Error("clients map should be initialised")
	}
	if tr.states == nil {
		t.Error("states map should be initialised")
	}
}

// ── Init ─────────────────────────────────────────────────────────────────────

func TestInit_CreatesState(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("USERS", 200)

	tr.mu.Lock()
	state, ok := tr.states["USERS"]
	tr.mu.Unlock()

	if !ok {
		t.Fatal("expected state to be created for USERS")
	}
	if state.total != 200 {
		t.Errorf("state.total = %d, want 200", state.total)
	}
	if state.lastCount != 0 {
		t.Errorf("state.lastCount should start at 0, got %d", state.lastCount)
	}
}

// ── Update throttling ─────────────────────────────────────────────────────────

func TestUpdate_FiresWhenTimeElapsed(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 1000)

	// Push lastTime far into the past so the 200 ms window has elapsed
	tr.mu.Lock()
	state := tr.states["T"]
	state.lastTime = time.Now().Add(-time.Second)
	tr.mu.Unlock()

	tr.Update("T", 50)

	tr.mu.Lock()
	got := state.lastCount
	tr.mu.Unlock()

	if got != 50 {
		t.Errorf("lastCount = %d after time-elapsed update, want 50", got)
	}
}

func TestUpdate_ThrottledWithinWindow(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 1000)

	// Set lastTime to now and lastCount to 10 — within the 200 ms window
	tr.mu.Lock()
	state := tr.states["T"]
	state.lastTime = time.Now()
	state.lastCount = 10
	tr.mu.Unlock()

	// 20 − 10 = 10 rows / 1000 total = 1% < 5% threshold
	tr.Update("T", 20)

	tr.mu.Lock()
	got := state.lastCount
	tr.mu.Unlock()

	if got != 10 {
		t.Errorf("lastCount = %d; update should have been throttled (want 10)", got)
	}
}

func TestUpdate_FiresOnSignificantProgress(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 100)

	tr.mu.Lock()
	state := tr.states["T"]
	state.lastTime = time.Now() // just now, so time won't trigger
	state.lastCount = 0
	tr.mu.Unlock()

	// 10 / 100 = 10% ≥ 5% — should send regardless of time
	tr.Update("T", 10)

	tr.mu.Lock()
	got := state.lastCount
	tr.mu.Unlock()

	if got != 10 {
		t.Errorf("lastCount = %d; significant-progress update should have fired (want 10)", got)
	}
}

func TestUpdate_ZeroTotalNeverTreatedAsSignificant(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 0) // total unknown

	tr.mu.Lock()
	state := tr.states["T"]
	state.lastTime = time.Now()
	state.lastCount = 0
	tr.mu.Unlock()

	// With total=0 progressSignificant is always false; only time triggers
	tr.Update("T", 5)

	tr.mu.Lock()
	got := state.lastCount
	tr.mu.Unlock()

	// Should NOT have updated (time hasn't elapsed, total=0)
	if got != 0 {
		t.Errorf("lastCount = %d; update should be throttled when total=0 and time hasn't elapsed (want 0)", got)
	}
}

func TestUpdate_UnknownTable_DoesNotPanic(t *testing.T) {
	tr := NewWebSocketTracker()
	// No Init call — should just return without panicking
	tr.Update("NONEXISTENT", 42)
}

// ── Done ──────────────────────────────────────────────────────────────────────

func TestDone_RemovesState(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 50)
	tr.Done("T")

	tr.mu.Lock()
	_, ok := tr.states["T"]
	tr.mu.Unlock()

	if ok {
		t.Error("expected state to be removed after Done")
	}
}

// ── Error ─────────────────────────────────────────────────────────────────────

func TestError_RemovesState(t *testing.T) {
	tr := NewWebSocketTracker()
	tr.Init("T", 50)
	tr.Error("T", errors.New("something went wrong"))

	tr.mu.Lock()
	_, ok := tr.states["T"]
	tr.mu.Unlock()

	if ok {
		t.Error("expected state to be removed after Error")
	}
}

// ── Full-stack WebSocket integration ─────────────────────────────────────────

// dialTestServer creates a Gin test server with HandleConnection registered and
// returns a connected WebSocket client plus a cleanup function.
func dialTestServer(t *testing.T, tr *WebSocketTracker) (*websocket.Conn, func()) {
	t.Helper()

	r := gin.New()
	r.GET("/ws", tr.HandleConnection)
	srv := httptest.NewServer(r)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("failed to dial WebSocket: %v", err)
	}

	// Wait until the server goroutine has registered the client
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tr.mu.Lock()
		n := len(tr.clients)
		tr.mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	return conn, func() {
		conn.Close()
		srv.Close()
	}
}

func TestHandleConnection_ReceivesInitMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("ORDERS", 500)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}

	if msg.Type != MsgInit {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgInit)
	}
	if msg.Table != "ORDERS" {
		t.Errorf("msg.Table = %q, want %q", msg.Table, "ORDERS")
	}
	if msg.Total != 500 {
		t.Errorf("msg.Total = %d, want 500", msg.Total)
	}
}

func TestHandleConnection_ReceivesDoneMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("T", 10)

	// Drain the init message first
	conn.SetReadDeadline(time.Now().Add(time.Second))
	var init ProgressMsg
	if err := conn.ReadJSON(&init); err != nil {
		t.Fatalf("failed to read init message: %v", err)
	}

	tr.Done("T")

	conn.SetReadDeadline(time.Now().Add(time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read done message: %v", err)
	}
	if msg.Type != MsgDone {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDone)
	}
}

func TestHandleConnection_ReceivesErrorMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("T", 10)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	var init ProgressMsg
	_ = conn.ReadJSON(&init)

	tr.Error("T", errors.New("disk full"))

	conn.SetReadDeadline(time.Now().Add(time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read error message: %v", err)
	}
	if msg.Type != MsgError {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgError)
	}
	if msg.ErrorMsg != "disk full" {
		t.Errorf("msg.ErrorMsg = %q, want %q", msg.ErrorMsg, "disk full")
	}
}

func TestHandleConnection_ReceivesAllDoneMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.AllDone("migration_20240101.zip")

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read all_done message: %v", err)
	}
	if msg.Type != MsgAllDone {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgAllDone)
	}
	if msg.ZipFileID != "migration_20240101.zip" {
		t.Errorf("msg.ZipFileID = %q, want %q", msg.ZipFileID, "migration_20240101.zip")
	}
}

func TestHandleConnection_ReceivesDryRunResultMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DryRunResult("ORDERS", 1234, true)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read dry_run_result message: %v", err)
	}
	if msg.Type != MsgDryRunResult {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDryRunResult)
	}
	if msg.Table != "ORDERS" {
		t.Errorf("msg.Table = %q, want %q", msg.Table, "ORDERS")
	}
	if msg.Total != 1234 {
		t.Errorf("msg.Total = %d, want 1234", msg.Total)
	}
	if !msg.ConnectionOk {
		t.Error("msg.ConnectionOk should be true")
	}
}

func TestHandleConnection_DryRunResult_ConnectionFailed(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DryRunResult("BAD_TABLE", 0, false)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read message: %v", err)
	}
	if msg.Type != MsgDryRunResult {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDryRunResult)
	}
	if msg.ConnectionOk {
		t.Error("msg.ConnectionOk should be false")
	}
}

func TestHandleConnection_ReceivesDDLProgressMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DDLProgress("sequence", "USERS_SEQ", "ok", nil)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read ddl_progress message: %v", err)
	}
	if msg.Type != MsgDDLProgress {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDDLProgress)
	}
	if msg.Object != "sequence" {
		t.Errorf("msg.Object = %q, want %q", msg.Object, "sequence")
	}
	if msg.ObjectName != "USERS_SEQ" {
		t.Errorf("msg.ObjectName = %q, want %q", msg.ObjectName, "USERS_SEQ")
	}
	if msg.Status != "ok" {
		t.Errorf("msg.Status = %q, want %q", msg.Status, "ok")
	}
	if msg.ErrorMsg != "" {
		t.Errorf("msg.ErrorMsg should be empty for ok status, got %q", msg.ErrorMsg)
	}
}

func TestDDLProgress_ErrorIncludesMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DDLProgress("index", "IDX_USERS_EMAIL", "error", errors.New("constraint violation"))

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read ddl_progress message: %v", err)
	}
	if msg.Type != MsgDDLProgress {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDDLProgress)
	}
	if msg.Object != "index" {
		t.Errorf("msg.Object = %q, want %q", msg.Object, "index")
	}
	if msg.Status != "error" {
		t.Errorf("msg.Status = %q, want %q", msg.Status, "error")
	}
	if msg.ErrorMsg != "constraint violation" {
		t.Errorf("msg.ErrorMsg = %q, want %q", msg.ErrorMsg, "constraint violation")
	}
}

func TestWarning(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Warning("This is a warning")

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg ProgressMsg
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("failed to read warning message: %v", err)
	}
	if msg.Type != MsgWarning {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgWarning)
	}
	if msg.Message != "This is a warning" {
		t.Errorf("msg.Message = %q, want %q", msg.Message, "This is a warning")
	}
}

func TestBroadcast_RemovesDeadClient(t *testing.T) {
	tr := NewWebSocketTracker()
	conn, cleanup := dialTestServer(t, tr)
	// Close the connection before the broadcast so the write will fail
	conn.Close()
	cleanup()

	// Give the server goroutine time to notice the close and remove the client
	time.Sleep(50 * time.Millisecond)

	// broadcast should not panic and should clean up the dead client
	tr.AllDone("")

	tr.mu.Lock()
	n := len(tr.clients)
	tr.mu.Unlock()
	if n != 0 {
		t.Errorf("expected 0 clients after dead-client cleanup, got %d", n)
	}
}
