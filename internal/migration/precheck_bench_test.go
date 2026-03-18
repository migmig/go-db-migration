package migration

import (
	"context"
	"fmt"
	"testing"
)

// BenchmarkRunPrecheckRowCount_1000Tables는 1,000개 테이블에 대한 병렬 pre-check 처리 시간을 측정한다.
func BenchmarkRunPrecheckRowCount_1000Tables(b *testing.B) {
	tables := make([]string, 1000)
	counts := make(map[string]int, 1000)
	for i := range tables {
		name := fmt.Sprintf("TABLE_%04d", i)
		tables[i] = name
		counts[name] = i * 100
	}

	sourceFn := mockCountFn(counts, nil)
	// target: ~half equal, ~half different
	targetCounts := make(map[string]int, 1000)
	for i, name := range tables {
		if i%2 == 0 {
			targetCounts[name] = counts[name] // equal
		} else {
			targetCounts[name] = counts[name] - 1 // different
		}
	}
	targetFn := mockCountFn(targetCounts, nil)

	cfg := PrecheckEngineConfig{
		Concurrency: 8,
		TimeoutMs:   1000,
		Policy:      PolicySkipEqualRows,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results, _ := RunPrecheckRowCount(context.Background(), tables, sourceFn, targetFn, cfg)
		if len(results) != 1000 {
			b.Fatalf("unexpected result count: %d", len(results))
		}
	}
}

// BenchmarkRunPrecheckRowCount_Concurrency는 동시성 수에 따른 처리 성능 차이를 비교한다.
func BenchmarkRunPrecheckRowCount_Concurrency1(b *testing.B) {
	benchmarkWithConcurrency(b, 1)
}

func BenchmarkRunPrecheckRowCount_Concurrency4(b *testing.B) {
	benchmarkWithConcurrency(b, 4)
}

func BenchmarkRunPrecheckRowCount_Concurrency16(b *testing.B) {
	benchmarkWithConcurrency(b, 16)
}

func benchmarkWithConcurrency(b *testing.B, concurrency int) {
	b.Helper()
	const numTables = 200
	tables := make([]string, numTables)
	counts := make(map[string]int, numTables)
	for i := range tables {
		name := fmt.Sprintf("T_%03d", i)
		tables[i] = name
		counts[name] = i * 10
	}

	sourceFn := mockCountFn(counts, nil)
	targetFn := mockCountFn(counts, nil)

	cfg := PrecheckEngineConfig{
		Concurrency: concurrency,
		TimeoutMs:   1000,
		Policy:      PolicyStrict,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RunPrecheckRowCount(context.Background(), tables, sourceFn, targetFn, cfg)
	}
}
