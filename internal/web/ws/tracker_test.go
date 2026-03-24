package ws

import (
	"dbmigrator/internal/bus"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
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
// returns a message stream, WebSocket connection, and cleanup function.
func dialTestServer(t *testing.T, tr *WebSocketTracker) (<-chan ProgressMsg, *websocket.Conn, func()) {
	t.Helper()

	r := gin.New()
	r.GET("/ws", tr.HandleConnection)
	serverConn, clientConn := net.Pipe()
	ln := newSingleConnListener(serverConn)
	srv := &http.Server{Handler: r}
	go func() {
		_ = srv.Serve(ln)
	}()

	u, err := url.Parse("ws://pipe/ws")
	if err != nil {
		_ = srv.Close()
		t.Fatalf("failed to parse websocket URL: %v", err)
	}

	conn, _, err := websocket.NewClient(clientConn, u, nil, 1024, 1024)
	if err != nil {
		_ = srv.Close()
		t.Fatalf("failed to dial WebSocket: %v", err)
	}

	msgs := make(chan ProgressMsg, 16)
	go func() {
		defer close(msgs)
		for {
			var msg ProgressMsg
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			msgs <- msg
		}
	}()

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

	return msgs, conn, func() {
		_ = conn.Close()
		_ = srv.Close()
	}
}

func readTestMessage(t *testing.T, msgs <-chan ProgressMsg) ProgressMsg {
	t.Helper()

	select {
	case msg, ok := <-msgs:
		if !ok {
			t.Fatal("message stream closed unexpectedly")
		}
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for websocket message")
		return ProgressMsg{}
	}
}

type singleConnListener struct {
	conn   net.Conn
	mu     sync.Mutex
	closed bool
}

func newSingleConnListener(conn net.Conn) *singleConnListener {
	return &singleConnListener{conn: conn}
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil, net.ErrClosed
	}
	if l.conn == nil {
		return nil, io.EOF
	}
	conn := l.conn
	l.conn = nil
	return conn, nil
}

func (l *singleConnListener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.closed = true
	if l.conn != nil {
		_ = l.conn.Close()
		l.conn = nil
	}
	return nil
}

func (l *singleConnListener) Addr() net.Addr {
	return dummyAddr("pipe")
}

type dummyAddr string

func (a dummyAddr) Network() string { return string(a) }

func (a dummyAddr) String() string { return string(a) }

func TestHandleConnection_ReceivesInitMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("ORDERS", 500)

	msg := readTestMessage(t, msgs)

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
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("T", 10)

	// Drain the init message first
	_ = readTestMessage(t, msgs)

	tr.Done("T")

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgDone {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDone)
	}
}

func TestHandleConnection_ReceivesErrorMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Init("T", 10)

	_ = readTestMessage(t, msgs)

	tr.Error("T", errors.New("disk full"))

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgError {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgError)
	}
	if msg.ErrorMsg != "disk full" {
		t.Errorf("msg.ErrorMsg = %q, want %q", msg.ErrorMsg, "disk full")
	}
}

func TestHandleConnection_ReceivesAllDoneMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.AllDone("migration_20240101.zip", nil)

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgAllDone {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgAllDone)
	}
	if msg.ZipFileID != "migration_20240101.zip" {
		t.Errorf("msg.ZipFileID = %q, want %q", msg.ZipFileID, "migration_20240101.zip")
	}
}

func TestHandleConnection_ReceivesDryRunResultMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DryRunResult("ORDERS", 1234, true)

	msg := readTestMessage(t, msgs)
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
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DryRunResult("BAD_TABLE", 0, false)

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgDryRunResult {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgDryRunResult)
	}
	if msg.ConnectionOk {
		t.Error("msg.ConnectionOk should be false")
	}
}

func TestHandleConnection_ReceivesDDLProgressMessage(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DDLProgress("sequence", "USERS_SEQ", "ok", nil)

	msg := readTestMessage(t, msgs)
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
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.DDLProgress("index", "IDX_USERS_EMAIL", "error", errors.New("constraint violation"))

	msg := readTestMessage(t, msgs)
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
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.Warning("This is a warning")

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgWarning {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgWarning)
	}
	if msg.Message != "This is a warning" {
		t.Errorf("msg.Message = %q, want %q", msg.Message, "This is a warning")
	}
}

func TestEventBusRetryBroadcast(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.EventBus().Publish(bus.Event{
		Type:        bus.EventRetry,
		Table:       "USERS",
		Attempt:     2,
		MaxAttempts: 4,
		WaitSeconds: 3,
		Message:     "timeout",
	})

	msg := readTestMessage(t, msgs)
	if msg.Type != MsgRetry {
		t.Errorf("msg.Type = %v, want %v", msg.Type, MsgRetry)
	}
	if msg.Table != "USERS" {
		t.Errorf("msg.Table = %q, want %q", msg.Table, "USERS")
	}
	if msg.Attempt != 2 || msg.MaxAttempts != 4 || msg.WaitSeconds != 3 {
		t.Errorf("unexpected retry payload: %+v", msg)
	}
	if msg.Message != "timeout" {
		t.Errorf("msg.Message = %q, want %q", msg.Message, "timeout")
	}
}

func TestEventBusSubscriptions(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, _, cleanup := dialTestServer(t, tr)
	defer cleanup()

	tr.EventBus().Publish(bus.Event{Type: bus.EventInit, Table: "USERS", Total: 10})
	msg := readTestMessage(t, msgs)
	if msg.Type != MsgInit || msg.Table != "USERS" || msg.Total != 10 {
		t.Fatalf("unexpected init message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventUpdate, Table: "USERS", Count: 3})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgUpdate || msg.Table != "USERS" || msg.Count != 3 {
		t.Fatalf("unexpected update message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventDone, Table: "USERS"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgDone || msg.Table != "USERS" {
		t.Fatalf("unexpected done message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventError, Table: "USERS", Error: errors.New("boom")})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgError || msg.Table != "USERS" || msg.ErrorMsg != "boom" {
		t.Fatalf("unexpected error message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventAllDone, ZipFileID: "x.zip"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgAllDone || msg.ZipFileID != "x.zip" {
		t.Fatalf("unexpected all_done message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventDryRunResult, Table: "ORDERS", Total: 9, ConnectionOk: true})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgDryRunResult || msg.Table != "ORDERS" || msg.Total != 9 || !msg.ConnectionOk {
		t.Fatalf("unexpected dry_run_result message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventDDLProgress, Object: "sequence", ObjectName: "SEQ_1", Status: "ok"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgDDLProgress || msg.Object != "sequence" || msg.ObjectName != "SEQ_1" || msg.Status != "ok" {
		t.Fatalf("unexpected ddl_progress message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventWarning, Message: "careful"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgWarning || msg.Message != "careful" {
		t.Fatalf("unexpected warning message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventValidationStart, Table: "USERS"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgValidationStart || msg.Table != "USERS" {
		t.Fatalf("unexpected validation_start message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventValidationResult, Table: "USERS", Total: 4, Count: 3, Status: "ok", Message: "done"})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgValidationResult || msg.Table != "USERS" || msg.Total != 4 || msg.Count != 3 || msg.Status != "ok" || msg.Message != "done" {
		t.Fatalf("unexpected validation_result message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventDiscoverySummary, ObjectGroup: "all", Tables: []string{"A"}, Sequences: []string{"S"}})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgDiscoverySummary || msg.ObjectGroup != "all" || len(msg.Tables) != 1 || msg.Tables[0] != "A" || len(msg.Sequences) != 1 || msg.Sequences[0] != "S" {
		t.Fatalf("unexpected discovery_summary message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventPartialSuccess, Table: "USERS", SkippedBatches: 2, EstimatedSkippedRows: 14})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgPartialSuccess || msg.Table != "USERS" || msg.SkippedBatches != 2 || msg.EstimatedSkippedRows != 14 {
		t.Fatalf("unexpected partial_success message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{
		Type:        bus.EventRetry,
		Table:       "USERS",
		Attempt:     2,
		MaxAttempts: 4,
		WaitSeconds: 3,
		Message:     "retrying",
	})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgRetry || msg.Table != "USERS" || msg.Attempt != 2 || msg.MaxAttempts != 4 || msg.WaitSeconds != 3 || msg.Message != "retrying" {
		t.Fatalf("unexpected retry message: %+v", msg)
	}

	tr.EventBus().Publish(bus.Event{Type: bus.EventMetrics, Message: `{"count":1}`})
	msg = readTestMessage(t, msgs)
	if msg.Type != MsgType(bus.EventMetrics) || msg.Message != `{"count":1}` {
		t.Fatalf("unexpected metrics message: %+v", msg)
	}
}

func TestBroadcast_RemovesDeadClient(t *testing.T) {
	tr := NewWebSocketTracker()
	msgs, conn, cleanup := dialTestServer(t, tr)
	_ = msgs
	// Close the connection before the broadcast so the write will fail
	conn.Close()
	cleanup()

	// Give the server goroutine time to notice the close and remove the client
	time.Sleep(50 * time.Millisecond)

	// broadcast should not panic and should clean up the dead client
	tr.AllDone("", nil)

	tr.mu.Lock()
	n := len(tr.clients)
	tr.mu.Unlock()
	if n != 0 {
		t.Errorf("expected 0 clients after dead-client cleanup, got %d", n)
	}
}
