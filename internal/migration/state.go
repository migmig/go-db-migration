package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type TableState struct {
	LastPKValue interface{} `json:"last_pk_value,omitempty"`
	Offset      int         `json:"offset,omitempty"`
	Completed   bool        `json:"completed"`
}

type MigrationState struct {
	JobID  string                 `json:"job_id"`
	Tables map[string]*TableState `json:"tables"`
	mu     sync.Mutex
}

func NewMigrationState(jobID string) *MigrationState {
	return &MigrationState{
		JobID:  jobID,
		Tables: make(map[string]*TableState),
	}
}

func LoadState(jobID string) (*MigrationState, error) {
	path := filepath.Join(".migration_state", fmt.Sprintf("%s.json", jobID))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewMigrationState(jobID), nil
		}
		return nil, err
	}

	state := &MigrationState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	if state.Tables == nil {
		state.Tables = make(map[string]*TableState)
	}
	state.JobID = jobID
	return state, nil
}

func (s *MigrationState) Save() error {
	dir := ".migration_state"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s.json", s.JobID))
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *MigrationState) UpdateOffset(table string, offset int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Tables[table]; !ok {
		s.Tables[table] = &TableState{}
	}
	s.Tables[table].Offset = offset
	s.Save()
}

func (s *MigrationState) MarkCompleted(table string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Tables[table]; !ok {
		s.Tables[table] = &TableState{}
	}
	s.Tables[table].Completed = true
	s.Save()
}

func (s *MigrationState) GetState(table string) *TableState {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ts, ok := s.Tables[table]; ok {
		// Return a copy to avoid data races when reading
		return &TableState{
			LastPKValue: ts.LastPKValue,
			Offset:      ts.Offset,
			Completed:   ts.Completed,
		}
	}
	return &TableState{}
}
