package ws

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local tool
	},
}

type MsgType string

const (
	MsgInit         MsgType = "init"
	MsgUpdate       MsgType = "update"
	MsgDone         MsgType = "done"
	MsgError        MsgType = "error"
	MsgAllDone      MsgType = "all_done"
	MsgDryRunResult MsgType = "dry_run_result"
	MsgDDLProgress  MsgType = "ddl_progress"
	MsgWarning      MsgType = "warning"
)

type ProgressMsg struct {
	Type         MsgType `json:"type"`
	Table        string  `json:"table,omitempty"`
	Count        int     `json:"count,omitempty"`
	Total        int     `json:"total,omitempty"`
	ErrorMsg     string  `json:"error,omitempty"`
	Message      string  `json:"message,omitempty"`
	ZipFileID    string  `json:"zip_file_id,omitempty"`
	ConnectionOk bool    `json:"connection_ok,omitempty"`
	Object       string  `json:"object,omitempty"`
	ObjectName   string  `json:"object_name,omitempty"`
	Status       string  `json:"status,omitempty"`
}

type tableState struct {
	total     int
	lastCount int
	lastTime  time.Time
}

type WebSocketTracker struct {
	clients map[*websocket.Conn]bool
	states  map[string]*tableState
	mu      sync.Mutex
}

func NewWebSocketTracker() *WebSocketTracker {
	return &WebSocketTracker{
		clients: make(map[*websocket.Conn]bool),
		states:  make(map[string]*tableState),
	}
}

func (t *WebSocketTracker) HandleConnection(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade error: %v", err)
		return
	}

	t.mu.Lock()
	t.clients[ws] = true
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.clients, ws)
		t.mu.Unlock()
		ws.Close()
	}()

	// Keep connection alive and read messages (though we mostly write)
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			break
		}
	}
}

func (t *WebSocketTracker) broadcast(msg ProgressMsg) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for client := range t.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(t.clients, client)
		}
	}
}

func (t *WebSocketTracker) Init(table string, totalRows int) {
	t.mu.Lock()
	t.states[table] = &tableState{
		total:     totalRows,
		lastCount: 0,
		lastTime:  time.Now(),
	}
	t.mu.Unlock()

	t.broadcast(ProgressMsg{
		Type:  MsgInit,
		Table: table,
		Total: totalRows,
	})
}

func (t *WebSocketTracker) Update(table string, processedRows int) {
	t.mu.Lock()
	state, exists := t.states[table]
	if !exists {
		t.mu.Unlock()
		return
	}

	now := time.Now()
	// Throttle: Send only if 200ms passed OR processed more than 5% of total since last update
	timeElapsed := now.Sub(state.lastTime) > 200*time.Millisecond
	progressSignificant := false
	if state.total > 0 {
		diff := processedRows - state.lastCount
		if float64(diff)/float64(state.total) >= 0.05 {
			progressSignificant = true
		}
	}

	shouldSend := timeElapsed || progressSignificant

	if shouldSend {
		state.lastCount = processedRows
		state.lastTime = now
	}
	t.mu.Unlock()

	if shouldSend {
		t.broadcast(ProgressMsg{
			Type:  MsgUpdate,
			Table: table,
			Count: processedRows,
		})
	}
}

func (t *WebSocketTracker) Done(table string) {
	t.mu.Lock()
	delete(t.states, table)
	t.mu.Unlock()

	t.broadcast(ProgressMsg{
		Type:  MsgDone,
		Table: table,
	})
}

func (t *WebSocketTracker) Error(table string, err error) {
	t.mu.Lock()
	delete(t.states, table)
	t.mu.Unlock()

	t.broadcast(ProgressMsg{
		Type:     MsgError,
		Table:    table,
		ErrorMsg: err.Error(),
	})
}

func (t *WebSocketTracker) AllDone(zipFileID string) {
	t.broadcast(ProgressMsg{
		Type:      MsgAllDone,
		ZipFileID: zipFileID,
	})
}

func (t *WebSocketTracker) DryRunResult(table string, totalRows int, connectionOk bool) {
	t.broadcast(ProgressMsg{
		Type:         MsgDryRunResult,
		Table:        table,
		Total:        totalRows,
		ConnectionOk: connectionOk,
	})
}

func (t *WebSocketTracker) DDLProgress(object, name, status string, err error) {
	msg := ProgressMsg{
		Type:       MsgDDLProgress,
		Object:     object,
		ObjectName: name,
		Status:     status,
	}
	if err != nil {
		msg.ErrorMsg = err.Error()
	}
	t.broadcast(msg)
}

func (t *WebSocketTracker) Warning(message string) {
	t.broadcast(ProgressMsg{
		Type:    MsgWarning,
		Message: message,
	})
}
