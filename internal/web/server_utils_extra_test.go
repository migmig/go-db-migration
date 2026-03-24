package web

import (
	"testing"
)

func TestBuildReplayPayload(t *testing.T) {
	req := startMigrationRequest{
		OracleURL: "u",
		Username:  "un",
		Password:  "p",
		TargetDB:  "pg",
		TargetURL: "purl",
		Tables:    []string{"T1"},
		Direct:    true,
		SessionID: "job",
		ObjectGroup: "all",
		OnError:   "fail_fast",
		BatchSize: 100,
		DBMaxOpen: 10,
		DBMaxIdle: 2,
		DBMaxLife: 0,
	}
	p := buildReplayPayload(req)
	if len(p) == 0 {
		t.Error("expected non-empty payload")
	}
}
