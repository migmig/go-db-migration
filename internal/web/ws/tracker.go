package ws

import (
	"log"
	"net/http"
	"sync"

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
	MsgInit   MsgType = "init"
	MsgUpdate MsgType = "update"
	MsgDone   MsgType = "done"
	MsgError  MsgType = "error"
	MsgAllDone MsgType = "all_done"
)

type ProgressMsg struct {
	Type      MsgType `json:"type"`
	Table     string  `json:"table,omitempty"`
	Count     int     `json:"count,omitempty"`
	Total     int     `json:"total,omitempty"`
	ErrorMsg  string  `json:"error,omitempty"`
	ZipFileID string  `json:"zip_file_id,omitempty"`
}

type WebSocketTracker struct {
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewWebSocketTracker() *WebSocketTracker {
	return &WebSocketTracker{
		clients: make(map[*websocket.Conn]bool),
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
	t.broadcast(ProgressMsg{
		Type:  MsgInit,
		Table: table,
		Total: totalRows,
	})
}

func (t *WebSocketTracker) Update(table string, processedRows int) {
	t.broadcast(ProgressMsg{
		Type:  MsgUpdate,
		Table: table,
		Count: processedRows,
	})
}

func (t *WebSocketTracker) Done(table string) {
	t.broadcast(ProgressMsg{
		Type:  MsgDone,
		Table: table,
	})
}

func (t *WebSocketTracker) Error(table string, err error) {
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
