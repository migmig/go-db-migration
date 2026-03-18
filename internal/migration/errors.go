package migration

import (
	"fmt"
	"strings"
)

// ErrorCategory는 마이그레이션 에러의 분류 코드이다.
type ErrorCategory string

const (
	ErrTypeMismatch     ErrorCategory = "TYPE_MISMATCH"
	ErrNullViolation    ErrorCategory = "NULL_VIOLATION"
	ErrUniqueViolation  ErrorCategory = "UNIQUE_VIOLATION"
	ErrFKViolation      ErrorCategory = "FK_VIOLATION"
	ErrConnectionLost   ErrorCategory = "CONNECTION_LOST"
	ErrTimeout          ErrorCategory = "TIMEOUT"
	ErrPermissionDenied ErrorCategory = "PERMISSION_DENIED"
	ErrUnknown          ErrorCategory = "UNKNOWN"
)

// MigrationError는 마이그레이션 과정에서 발생하는 구조화된 에러이다.
type MigrationError struct {
	Table       string
	Phase       string // "ddl", "data", "index", "constraint", "validation"
	Category    ErrorCategory
	BatchNum    int    // 1-based 배치 번호 (data phase에서만 유효)
	RowOffset   int    // 전체 행 기준 오프셋
	Column      string // 문제 컬럼 (파악 가능한 경우)
	RootCause   error
	Suggestion  string
	Recoverable bool
}

func (e *MigrationError) Error() string {
	msg := fmt.Sprintf("[%s] %s (table=%s, phase=%s", e.Category, e.RootCause, e.Table, e.Phase)
	if e.BatchNum > 0 {
		msg += fmt.Sprintf(", batch=%d", e.BatchNum)
	}
	if e.RowOffset > 0 {
		msg += fmt.Sprintf(", row=%d", e.RowOffset)
	}
	if e.Column != "" {
		msg += fmt.Sprintf(", column=%s", e.Column)
	}
	msg += ")"
	return msg
}

func (e *MigrationError) Unwrap() error {
	return e.RootCause
}

// RetryEvent는 자동 재시도 시 UI/관측성으로 전달되는 이벤트 payload이다.
type RetryEvent struct {
	TableName   string
	Attempt     int
	MaxAttempts int
	ErrorMsg    string
	WaitSeconds int
}

// PartialBatchError는 skip_batch 정책으로 일부 배치를 건너뛰고 완료했을 때 사용된다.
type PartialBatchError struct {
	Table                string
	SkippedBatches       int
	EstimatedSkippedRows int
	Cause                error
}

func (e *PartialBatchError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("partial success (table=%s, skipped_batches=%d, estimated_skipped_rows=%d): %v",
			e.Table, e.SkippedBatches, e.EstimatedSkippedRows, e.Cause)
	}
	return fmt.Sprintf("partial success (table=%s, skipped_batches=%d, estimated_skipped_rows=%d)",
		e.Table, e.SkippedBatches, e.EstimatedSkippedRows)
}

// DetailedError는 ws 패키지가 순환 의존 없이 에러 상세 필드를 읽을 수 있도록 하는 인터페이스이다.
type DetailedError interface {
	ErrorPhase() string
	ErrorCategory() string
	ErrorSuggestion() string
	IsRecoverable() bool
}

func (e *MigrationError) ErrorPhase() string      { return e.Phase }
func (e *MigrationError) ErrorCategory() string   { return string(e.Category) }
func (e *MigrationError) ErrorSuggestion() string { return e.Suggestion }
func (e *MigrationError) IsRecoverable() bool     { return e.Recoverable }
func (e *MigrationError) ErrorBatchNum() int      { return e.BatchNum }
func (e *MigrationError) ErrorRowOffset() int     { return e.RowOffset }

// classifyError는 DB 드라이버 에러 메시지를 분석하여 ErrorCategory를 결정한다.
func classifyError(err error) ErrorCategory {
	if err == nil {
		return ErrUnknown
	}
	msg := err.Error()
	switch {
	case containsAny(msg, "data type", "type mismatch", "incompatible", "too long", "overflow", "value too large", "ORA-01401", "ORA-01438", "ORA-12899"):
		return ErrTypeMismatch
	case containsAny(msg, "cannot insert NULL", "NOT NULL constraint", "null value in column", "ORA-01400", "Column cannot be null"):
		return ErrNullViolation
	case containsAny(msg, "foreign key", "referential integrity", "REFERENCES", "ORA-02291", "ORA-02292"):
		return ErrFKViolation
	case containsAny(msg, "connection refused", "connection reset", "broken pipe", "EOF", "no route to host", "i/o timeout", "network"):
		return ErrConnectionLost
	case containsAny(msg, "timeout", "deadline exceeded", "context deadline", "ORA-01013"):
		return ErrTimeout
	case containsAny(msg, "permission denied", "insufficient privileges", "ORA-01031", "Access denied"):
		return ErrPermissionDenied
	default:
		return ErrUnknown
	}
}

// containsAny는 s가 substrs 중 하나라도 포함하는지 확인한다 (대소문자 무시).
func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// suggestFix는 에러 카테고리와 타겟 dialect 이름 기반으로 복구 제안 메시지를 반환한다.
func suggestFix(category ErrorCategory, dialectName string) string {
	switch category {
	case ErrTypeMismatch:
		switch dialectName {
		case "mysql", "mariadb":
			return "Check column types: CLOB→LONGTEXT, VARCHAR2 length may exceed target column size. Review column definitions."
		case "mssql":
			return "Check column types: CLOB→NVARCHAR(MAX), ensure DECIMAL precision matches. Review column definitions."
		default:
			return "Check column type mappings between Oracle and target DB. Consider using --with-ddl to auto-create tables."
		}
	case ErrNullViolation:
		return "A NOT NULL column received a NULL value. Check Oracle source data or remove NOT NULL constraint on target before migration."
	case ErrFKViolation:
		return "Foreign key constraint violation. Ensure referenced tables are migrated first, or disable FK checks during migration."
	case ErrConnectionLost:
		return "Connection lost. Check network stability, increase timeout settings, and use --resume to continue from checkpoint."
	case ErrTimeout:
		return "Query timed out. Consider reducing --batch size, increasing --db-max-life, or splitting large tables."
	case ErrPermissionDenied:
		return "Insufficient database privileges. Verify GRANT SELECT on Oracle source and INSERT/CREATE on target DB."
	default:
		return "Check logs for details. Use --log-json for structured error output."
	}
}
