package ws

import (
	"log"
	"net/http"
	"sync"
	"time"

	"dbmigrator/internal/bus"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local tool
	},
}

type MsgType string

const (
	MsgInit             MsgType = "init"
	MsgUpdate           MsgType = "update"
	MsgDone             MsgType = "done"
	MsgError            MsgType = "error"
	MsgAllDone          MsgType = "all_done"
	MsgDryRunResult     MsgType = "dry_run_result"
	MsgDDLProgress      MsgType = "ddl_progress"
	MsgWarning          MsgType = "warning"
	MsgValidationStart  MsgType = "validation_start"
	MsgValidationResult MsgType = "validation_result"
	MsgDiscoverySummary MsgType = "discovery_summary"
	MsgMetrics          MsgType = "metrics"
)

// DetailedError는 migration 패키지의 순환 의존 없이 구조화 에러 필드를 읽기 위한 인터페이스이다.
type DetailedError interface {
	ErrorPhase() string
	ErrorCategory() string
	ErrorSuggestion() string
	IsRecoverable() bool
	ErrorBatchNum() int
	ErrorRowOffset() int
}

// ReportSummary는 마이그레이션 완료 시 all_done 메시지에 포함되는 요약 정보이다.
type ReportSummary struct {
	TotalRows    int          `json:"total_rows"`
	SuccessCount int          `json:"success_count"`
	ErrorCount   int          `json:"error_count"`
	Duration     string       `json:"duration"`
	ReportID     string       `json:"report_id"`
	ObjectGroup  string       `json:"object_group"`
	Stats        GroupedStats `json:"stats"`
}

type GroupStats struct {
	TotalItems   int `json:"total_items"`
	SuccessCount int `json:"success_count"`
	ErrorCount   int `json:"error_count"`
	SkippedCount int `json:"skipped_count"`
	TotalRows    int `json:"total_rows,omitempty"`
}

type GroupedStats struct {
	Tables    GroupStats `json:"tables"`
	Sequences GroupStats `json:"sequences"`
}

type ProgressMsg struct {
	Type         MsgType  `json:"type"`
	Table        string   `json:"table,omitempty"`
	Count        int      `json:"count,omitempty"`
	Total        int      `json:"total,omitempty"`
	ErrorMsg     string   `json:"error,omitempty"`
	Message      string   `json:"message,omitempty"`
	ZipFileID    string   `json:"zip_file_id,omitempty"`
	ConnectionOk bool     `json:"connection_ok,omitempty"`
	Object       string   `json:"object,omitempty"`
	ObjectName   string   `json:"object_name,omitempty"`
	Status       string   `json:"status,omitempty"`
	ObjectGroup  string   `json:"object_group,omitempty"`
	Tables       []string `json:"tables,omitempty"`
	Sequences    []string `json:"sequences,omitempty"`
	// v9: 구조화 에러 필드
	Phase       string `json:"phase,omitempty"`
	Category    string `json:"category,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
	Recoverable *bool  `json:"recoverable,omitempty"`
	BatchNum    int    `json:"batch_num,omitempty"`
	RowOffset   int    `json:"row_offset,omitempty"`
	// v9: 리포트 요약
	ReportSummary *ReportSummary `json:"report_summary,omitempty"`
}

type tableState struct {
	total     int
	lastCount int
	lastTime  time.Time
}

type WebSocketTracker struct {
	clients      map[*websocket.Conn]bool
	states       map[string]*tableState
	mu           sync.Mutex
	lastAccessed time.Time
	eventBus     bus.EventBus
}

func NewWebSocketTracker() *WebSocketTracker {
	t := &WebSocketTracker{
		clients:      make(map[*websocket.Conn]bool),
		states:       make(map[string]*tableState),
		lastAccessed: time.Now(),
		eventBus:     bus.NewEventBus(),
	}
	t.setupSubscriptions()
	return t
}

func (t *WebSocketTracker) setupSubscriptions() {
	t.eventBus.Subscribe(bus.EventInit, func(e bus.Event) { t.Init(e.Table, e.Total) })
	t.eventBus.Subscribe(bus.EventUpdate, func(e bus.Event) { t.Update(e.Table, e.Count) })
	t.eventBus.Subscribe(bus.EventDone, func(e bus.Event) { t.Done(e.Table) })
	t.eventBus.Subscribe(bus.EventError, func(e bus.Event) { t.Error(e.Table, e.Error) })
	t.eventBus.Subscribe(bus.EventAllDone, func(e bus.Event) {
		var summary *ReportSummary
		if e.ReportSummary != nil {
			if s, ok := e.ReportSummary.(*ReportSummary); ok {
				summary = s
			}
		}
		t.AllDone(e.ZipFileID, summary)
	})
	t.eventBus.Subscribe(bus.EventDryRunResult, func(e bus.Event) { t.DryRunResult(e.Table, e.Total, e.ConnectionOk) })
	t.eventBus.Subscribe(bus.EventDDLProgress, func(e bus.Event) { t.DDLProgress(e.Object, e.ObjectName, e.Status, e.Error) })
	t.eventBus.Subscribe(bus.EventWarning, func(e bus.Event) { t.Warning(e.Message) })
	t.eventBus.Subscribe(bus.EventValidationStart, func(e bus.Event) { t.ValidationStart(e.Table) })
	t.eventBus.Subscribe(bus.EventValidationResult, func(e bus.Event) { t.ValidationResult(e.Table, e.Total, e.Count, e.Status, e.Message) })
	t.eventBus.Subscribe(bus.EventDiscoverySummary, func(e bus.Event) { t.DiscoverySummary(e.ObjectGroup, e.Tables, e.Sequences) })
	t.eventBus.Subscribe(bus.EventMetrics, func(e bus.Event) {
		t.broadcast(ProgressMsg{
			Type:    MsgType(e.Type),
			Message: e.Message, // JSON payload representing metrics
		})
	})
}

// EventBus returns the underlying EventBus for publishing events
func (t *WebSocketTracker) EventBus() bus.EventBus {
	return t.eventBus
}

type SessionManager struct {
	trackers map[string]*WebSocketTracker
	mu       sync.Mutex
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		trackers: make(map[string]*WebSocketTracker),
	}
	go sm.cleanupLoop()
	return sm
}

func (sm *SessionManager) CreateSession() string {
	id := uuid.New().String()
	sm.mu.Lock()
	sm.trackers[id] = NewWebSocketTracker()
	sm.mu.Unlock()
	return id
}

func (sm *SessionManager) GetTracker(id string) *WebSocketTracker {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.trackers[id]
}

func (sm *SessionManager) HandleConnection(c *gin.Context) {
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing sessionId"})
		return
	}

	sm.mu.Lock()
	tracker, exists := sm.trackers[sessionID]
	if !exists {
		tracker = NewWebSocketTracker()
		sm.trackers[sessionID] = tracker
	}
	sm.mu.Unlock()

	tracker.HandleConnection(c)
}

func (sm *SessionManager) cleanupLoop() {
	for {
		time.Sleep(5 * time.Minute)
		sm.mu.Lock()
		now := time.Now()
		for id, tracker := range sm.trackers {
			tracker.mu.Lock()
			clientsCount := len(tracker.clients)
			lastAcc := tracker.lastAccessed
			tracker.mu.Unlock()

			if clientsCount == 0 && now.Sub(lastAcc) > 30*time.Minute {
				delete(sm.trackers, id)
			}
		}
		sm.mu.Unlock()
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
	t.lastAccessed = time.Now()
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		delete(t.clients, ws)
		t.lastAccessed = time.Now()
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

	msg := ProgressMsg{
		Type:     MsgError,
		Table:    table,
		ErrorMsg: err.Error(),
	}

	// MigrationError 등 DetailedError 구현체이면 상세 필드 추가
	if de, ok := err.(DetailedError); ok {
		msg.Phase = de.ErrorPhase()
		msg.Category = de.ErrorCategory()
		msg.Suggestion = de.ErrorSuggestion()
		recoverable := de.IsRecoverable()
		msg.Recoverable = &recoverable
		msg.BatchNum = de.ErrorBatchNum()
		msg.RowOffset = de.ErrorRowOffset()
	}

	t.broadcast(msg)
}

func (t *WebSocketTracker) AllDone(zipFileID string, summary *ReportSummary) {
	t.broadcast(ProgressMsg{
		Type:          MsgAllDone,
		ZipFileID:     zipFileID,
		ReportSummary: summary,
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

func (t *WebSocketTracker) ValidationStart(table string) {
	t.broadcast(ProgressMsg{
		Type:  MsgValidationStart,
		Table: table,
	})
}

func (t *WebSocketTracker) ValidationResult(table string, sourceCount, targetCount int, status, detail string) {
	t.broadcast(ProgressMsg{
		Type:    MsgValidationResult,
		Table:   table,
		Total:   sourceCount,
		Count:   targetCount,
		Status:  status,
		Message: detail,
	})
}

func (t *WebSocketTracker) DiscoverySummary(objectGroup string, tables, sequences []string) {
	t.broadcast(ProgressMsg{
		Type:        MsgDiscoverySummary,
		ObjectGroup: objectGroup,
		Tables:      tables,
		Sequences:   sequences,
	})
}
