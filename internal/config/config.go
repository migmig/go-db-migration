package config

import (
	"flag"
	"fmt"
	"strings"
)

type Config struct {
	OracleURL string
	User      string
	Password  string
	Tables    []string
	OutFile   string
	BatchSize int
	Schema    string
	PerTable  bool
	Parallel  bool
	// v2 flags
	PGURL   string
	Workers int
	WithDDL bool
	DryRun  bool
	LogJSON bool
	// v3 flags
	WebMode bool
	// v5 flags
	WithSequences bool
	WithIndexes   bool
	Sequences     string
	OracleOwner   string
	// v6 flags
	TargetDB  string
	TargetURL string
	// v8 flags
	WithConstraints bool
	DBMaxOpen       int
	DBMaxIdle       int
	DBMaxLife       int
	ResumeJobID     string
	// internal use for zip generation
	OutputDir string
}

func ParseFlags() (*Config, error) {
	cfg := &Config{}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "사용법: dbmigrator [옵션]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "옵션:\n")
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), `
사용 예시:
  # SQL 파일 생성 (Oracle → SQL 파일)
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS,ORDERS

  # 테이블별 SQL 파일 생성
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS,ORDERS -per-table -out export.sql

  # PostgreSQL 직접 마이그레이션
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS \
    -pg-url postgres://pguser:pgpass@localhost:5432/mydb

  # 스키마 지정 + DDL 포함
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS \
    -pg-url postgres://pguser:pgpass@localhost:5432/mydb -schema myschema -with-ddl

  # Dry-Run (연결 확인 및 행 수 추정만)
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS,ORDERS -dry-run

  # Sequence + Index DDL 포함
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS \
    -with-ddl -with-sequences -with-indexes

  # 소유자 명시 + Sequence 직접 지정
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS \
    -with-ddl -with-sequences -oracle-owner HR -sequences SEQ_USERS,SEQ_ORDERS

  # 제약조건(Default, FK, Check) 포함
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -tables USERS -with-ddl -with-constraints
  
  # 재개 기능 (실패한 마이그레이션 이어하기)
  dbmigrator -url localhost:1521/ORCL -user scott -password tiger -resume 20260313150405

  # Web UI 모드
  dbmigrator -web
`)
	}

	flag.StringVar(&cfg.OracleURL, "url", "", "Oracle 데이터베이스 URL (예: host:port/service_name)")
	flag.StringVar(&cfg.User, "user", "", "데이터베이스 사용자명")
	flag.StringVar(&cfg.Password, "password", "", "데이터베이스 비밀번호")
	tablesFlag := flag.String("tables", "", "마이그레이션할 테이블 목록 (쉼표로 구분)")
	flag.StringVar(&cfg.OutFile, "out", "migration.sql", "출력 SQL 파일명")
	flag.IntVar(&cfg.BatchSize, "batch", 1000, "INSERT 배치당 행 수")
	flag.StringVar(&cfg.Schema, "schema", "", "PostgreSQL 스키마 이름 (선택)")
	flag.BoolVar(&cfg.PerTable, "per-table", false, "테이블별 별도 파일로 출력")
	flag.BoolVar(&cfg.Parallel, "parallel", false, "테이블 병렬 처리")
	// v2 flags
	flag.StringVar(&cfg.PGURL, "pg-url", "", "PostgreSQL 연결 URL (예: postgres://user:pass@host:port/db)")
	flag.IntVar(&cfg.Workers, "workers", 4, "병렬 처리 워커 수")
	flag.BoolVar(&cfg.WithDDL, "with-ddl", false, "CREATE TABLE DDL 생성 포함")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "연결 확인 및 행 수 추정만 수행 (실제 마이그레이션 없음)")
	flag.BoolVar(&cfg.LogJSON, "log-json", false, "JSON 구조화 로그 활성화")
	flag.BoolVar(&cfg.WebMode, "web", false, "Web UI 모드로 실행")
	// v5 flags
	flag.BoolVar(&cfg.WithSequences, "with-sequences", false, "연관 Sequence DDL 포함")
	flag.BoolVar(&cfg.WithIndexes, "with-indexes", false, "연관 Index DDL 포함")
	flag.StringVar(&cfg.Sequences, "sequences", "", "추가 포함할 Sequence 이름 목록 (쉼표 구분)")
	flag.StringVar(&cfg.OracleOwner, "oracle-owner", "", "Oracle 스키마 소유자 (미지정 시 -user 값 사용)")
	// v6 flags
	flag.StringVar(&cfg.TargetDB, "target-db", "postgres", "출력 대상 DB 종류 (postgres/mysql/mariadb/sqlite/mssql)")
	flag.StringVar(&cfg.TargetURL, "target-url", "", "대상 DB 연결 URL (PostgreSQL 외 Direct 마이그레이션 시)")
	// v8 flags
	flag.BoolVar(&cfg.WithConstraints, "with-constraints", false, "제약조건(Default, FK, Check) 마이그레이션 포함")
	flag.IntVar(&cfg.DBMaxOpen, "db-max-open", 0, "DB 커넥션 풀 최대 활성 연결 수 (기본값: 0, 무제한)")
	flag.IntVar(&cfg.DBMaxIdle, "db-max-idle", 2, "DB 커넥션 풀 최대 유휴 연결 수 (기본값: 2)")
	flag.IntVar(&cfg.DBMaxLife, "db-max-life", 0, "DB 커넥션 풀 최대 유지 시간(초) (기본값: 0, 무제한)")
	flag.StringVar(&cfg.ResumeJobID, "resume", "", "재개할 Job ID")

	flag.Parse()

	// v6: Backward compatibility for pg-url
	if cfg.TargetDB == "postgres" && cfg.PGURL != "" {
		cfg.TargetURL = cfg.PGURL
	} else if cfg.TargetDB != "postgres" && cfg.PGURL != "" {
		fmt.Println("Warning: -pg-url is specified but -target-db is not postgres. -pg-url will be ignored.")
	}

	if cfg.WebMode {
		return cfg, nil
	}

	if cfg.OracleURL == "" || cfg.User == "" || cfg.Password == "" {
		flag.Usage()
		return nil, fmt.Errorf("missing required flags")
	}

	if *tablesFlag == "" && cfg.ResumeJobID == "" {
		flag.Usage()
		return nil, fmt.Errorf("missing required flags: -tables or -resume must be provided")
	}

	if *tablesFlag != "" {
		t := strings.Split(*tablesFlag, ",")
		for i := range t {
			cfg.Tables = append(cfg.Tables, strings.TrimSpace(t[i]))
		}
	}

	return cfg, nil
}
