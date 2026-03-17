package web

import (
	"sync"

	"dbmigrator/internal/migration"
)

// precheckResultStore는 마지막 pre-check 실행 결과를 인메모리에 보관한다.
type precheckResultStore struct {
	mu      sync.RWMutex
	results []migration.PrecheckTableResult
	summary migration.PrecheckSummary
}

func newPrecheckResultStore() *precheckResultStore {
	return &precheckResultStore{}
}

func (s *precheckResultStore) set(results []migration.PrecheckTableResult, summary migration.PrecheckSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.results = results
	s.summary = summary
}

func (s *precheckResultStore) getAll() ([]migration.PrecheckTableResult, migration.PrecheckSummary) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]migration.PrecheckTableResult, len(s.results))
	copy(out, s.results)
	return out, s.summary
}

// globalPrecheckStore는 서버 전체에서 공유하는 pre-check 결과 저장소다.
var globalPrecheckStore = newPrecheckResultStore()
