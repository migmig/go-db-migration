package config

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var osExit = os.Exit

func getParentProcessName() string {
	ppid := os.Getppid()
	out, err := exec.Command("ps", "-p", strconv.Itoa(ppid), "-o", "comm=").Output()
	if err == nil {
		comm := strings.TrimSpace(string(out))
		base := filepath.Base(comm)
		return strings.TrimPrefix(base, "-")
	}
	return ""
}

func detectShell() string {
	// 1. Try to detect from parent process name
	parent := getParentProcessName()
	if parent != "" {
		parent = strings.ToLower(parent)
		if strings.Contains(parent, "bash") {
			return "bash"
		}
		if strings.Contains(parent, "zsh") {
			return "zsh"
		}
		if strings.Contains(parent, "fish") {
			return "fish"
		}
		if strings.Contains(parent, "pwsh") || strings.Contains(parent, "powershell") {
			return "powershell"
		}
	}

	// 2. Fallback to $SHELL environment variable
	shellEnv := strings.ToLower(os.Getenv("SHELL"))
	if strings.Contains(shellEnv, "bash") {
		return "bash"
	}
	if strings.Contains(shellEnv, "zsh") {
		return "zsh"
	}
	if strings.Contains(shellEnv, "fish") {
		return "fish"
	}
	if strings.Contains(shellEnv, "pwsh") || strings.Contains(shellEnv, "powershell") {
		return "powershell"
	}
	return ""
}

func printCompletionUsage() {
	fmt.Fprintf(os.Stderr, "자동 감지된 쉘이 지원되지 않거나 알 수 없습니다.\n\n사용법:\n  -completion <shell>\n\n지원하는 쉘(shell):\n  bash, zsh, fish, powershell\n\n사용 예시:\n  ./dbmigrator -completion bash > /etc/bash_completion.d/dbmigrator\n  ./dbmigrator -completion zsh > ~/.zsh/completions/_dbmigrator\n")
}

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
	// v9 flags
	Validate  bool
	CopyBatch int
	// v12 flags
	CompletionShell string
	// v15 flags
	AuthEnabled bool
	MasterKey   string
}

func generateCompletionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return `# bash completion for dbmigrator
_dbmigrator_completions()
{
    local cur
    cur="${COMP_WORDS[COMP_CWORD]}"

    local opts="-web -url -user -password -tables -target-db -target-url -pg-url -out -batch -schema -per-table -parallel -workers -with-ddl -with-sequences -with-indexes -with-constraints -sequences -oracle-owner -db-max-open -db-max-idle -db-max-life -validate -copy-batch -resume -dry-run -log-json -completion -auth-enabled"
    local target_dbs="postgres mysql mariadb sqlite mssql"

    case "${COMP_WORDS[COMP_CWORD-1]}" in
        -completion)
            COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- "$cur") )
            return 0
            ;;
        -target-db)
            COMPREPLY=( $(compgen -W "$target_dbs" -- "$cur") )
            return 0
            ;;
    esac

    COMPREPLY=( $(compgen -W "$opts" -- "$cur") )
}

complete -F _dbmigrator_completions dbmigrator
`, nil
	case "zsh":
		return `#compdef dbmigrator

_dbmigrator() {
  local -a opts
  opts=(
    '-web[Web UI 모드로 실행]'
    '-url[Oracle 데이터베이스 URL]:url:'
    '-user[Oracle 데이터베이스 사용자명]:user:'
    '-password[Oracle 데이터베이스 비밀번호]:password:'
    '-tables[마이그레이션할 테이블 목록]:tables:'
    '-target-db[대상 DB 종류]:target db:(postgres mysql mariadb sqlite mssql)'
    '-target-url[대상 DB 연결 URL]:url:'
    '-pg-url[PostgreSQL 연결 URL]:url:'
    '-out[출력 SQL 파일명]:filename:_files'
    '-batch[INSERT 배치당 행 수]:batch size:'
    '-schema[PostgreSQL 스키마 이름]:schema:'
    '-per-table[테이블별 별도 SQL 파일로 출력]'
    '-parallel[테이블 병렬 처리]'
    '-workers[병렬 처리 워커 수]:workers:'
    '-with-ddl[CREATE TABLE DDL 생성 포함]'
    '-with-sequences[연관 Sequence DDL 포함]'
    '-with-indexes[연관 Index DDL 포함]'
    '-with-constraints[제약조건 마이그레이션 포함]'
    '-sequences[추가 포함할 Sequence 목록]:sequences:'
    '-oracle-owner[Oracle 스키마 소유자]:owner:'
    '-db-max-open[최대 활성 연결 수]:count:'
    '-db-max-idle[최대 유휴 연결 수]:count:'
    '-db-max-life[최대 유지 시간(초)]:seconds:'
    '-validate[행 수 검증 수행]'
    '-copy-batch[PostgreSQL COPY 배치 크기]:batch:'
    '-resume[재개할 Job ID]:job id:'
    '-dry-run[연결 확인 및 행 수 추정만 수행]'
    '-log-json[JSON 구조화 로그 활성화]'
    '-completion[자동완성 스크립트 생성]:shell:(bash zsh fish powershell)'
    '-auth-enabled[인증 기반 멀티유저 모드 활성화]'
    )
    _arguments -s $opts
    }

    if [[ "$(basename -- ${(%):-%x})" != "_dbmigrator" ]]; then
    compdef _dbmigrator dbmigrator
    fi
    `, nil
	case "fish":
		return `# fish completion for dbmigrator
complete -c dbmigrator -f
complete -c dbmigrator -l web -d 'Web UI 모드로 실행'
complete -c dbmigrator -l url -r -d 'Oracle 데이터베이스 URL'
complete -c dbmigrator -l user -r -d 'Oracle 데이터베이스 사용자명'
complete -c dbmigrator -l password -r -d 'Oracle 데이터베이스 비밀번호'
complete -c dbmigrator -l tables -r -d '마이그레이션할 테이블 목록'
complete -c dbmigrator -l target-db -r -a 'postgres mysql mariadb sqlite mssql' -d '대상 DB 종류'
complete -c dbmigrator -l target-url -r -d '대상 DB 연결 URL'
complete -c dbmigrator -l pg-url -r -d 'PostgreSQL 연결 URL'
complete -c dbmigrator -l out -r -d '출력 SQL 파일명'
complete -c dbmigrator -l batch -r -d 'INSERT 배치당 행 수'
complete -c dbmigrator -l schema -r -d 'PostgreSQL 스키마 이름'
complete -c dbmigrator -l per-table -d '테이블별 별도 SQL 파일로 출력'
complete -c dbmigrator -l parallel -d '테이블 병렬 처리'
complete -c dbmigrator -l workers -r -d '병렬 처리 워커 수'
complete -c dbmigrator -l with-ddl -d 'CREATE TABLE DDL 생성 포함'
complete -c dbmigrator -l with-sequences -d '연관 Sequence DDL 포함'
complete -c dbmigrator -l with-indexes -d '연관 Index DDL 포함'
complete -c dbmigrator -l with-constraints -d '제약조건 마이그레이션 포함'
complete -c dbmigrator -l sequences -r -d '추가 포함할 Sequence 목록'
complete -c dbmigrator -l oracle-owner -r -d 'Oracle 스키마 소유자'
complete -c dbmigrator -l db-max-open -r -d '최대 활성 연결 수'
complete -c dbmigrator -l db-max-idle -r -d '최대 유휴 연결 수'
complete -c dbmigrator -l db-max-life -r -d '최대 유지 시간(초)'
complete -c dbmigrator -l validate -d '행 수 검증 수행'
complete -c dbmigrator -l copy-batch -r -d 'PostgreSQL COPY 배치 크기'
complete -c dbmigrator -l resume -r -d '재개할 Job ID'
complete -c dbmigrator -l dry-run -d '연결 확인 및 행 수 추정만 수행'
complete -c dbmigrator -l log-json -d 'JSON 구조화 로그 활성화'
complete -c dbmigrator -l auth-enabled -d '인증 기반 멀티유저 모드 활성화'
complete -c dbmigrator -l completion -r -a 'bash zsh fish powershell' -d '자동완성 스크립트 생성'
`, nil
	case "powershell":
		return `Register-ArgumentCompleter -Native -CommandName dbmigrator -ScriptBlock {
    param($wordToComplete)
    $opts = @(
        '-web','-url','-user','-password','-tables','-target-db','-target-url','-pg-url','-out','-batch',
        '-schema','-per-table','-parallel','-workers','-with-ddl','-with-sequences','-with-indexes',
        '-with-constraints','-sequences','-oracle-owner','-db-max-open','-db-max-idle','-db-max-life',
        '-validate','-copy-batch','-resume','-dry-run','-log-json','-completion','-auth-enabled'
    )
    $opts | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterName', $_)
    }
}
`, nil
	default:
		return "", fmt.Errorf("unsupported shell for -completion: %s (supported: bash, zsh, fish, powershell)", shell)
	}
}

func ParseFlags() (*Config, error) {
	// flag.Parse() 호출 전, -completion 단독 사용 예외 처리 및 쉘 자동 감지
	args := os.Args[1:]
	for i, arg := range args {
		if arg == "-completion" || arg == "--completion" {
			// 마지막 인자이거나 다음 인자가 또 다른 플래그일 때 (인자 없음)
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") {
				detected := detectShell()
				if detected != "" {
					script, _ := generateCompletionScript(detected)
					fmt.Println(script)
					osExit(0)
					return nil, nil // For tests
				}
				printCompletionUsage()
				osExit(1)
				return nil, nil // For tests
			}
		}
	}

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
	// v9 flags
	flag.BoolVar(&cfg.Validate, "validate", false, "마이그레이션 후 소스-타겟 행 수 검증 수행")
	flag.IntVar(&cfg.CopyBatch, "copy-batch", 10000, "PostgreSQL COPY 배치 크기 (0: 단일 COPY 모드)")
	// v12 flags
	flag.StringVar(&cfg.CompletionShell, "completion", "", "쉘 자동완성 스크립트 출력 (bash/zsh/fish/powershell)")
	flag.BoolVar(&cfg.AuthEnabled, "auth-enabled", false, "인증 기반 멀티유저 모드 활성화")

	flag.Parse()

	cfg.MasterKey = os.Getenv("DBM_MASTER_KEY")

	// v6: Backward compatibility for pg-url
	if cfg.TargetDB == "postgres" && cfg.PGURL != "" {
		cfg.TargetURL = cfg.PGURL
	} else if cfg.TargetDB != "postgres" && cfg.PGURL != "" {
		fmt.Println("Warning: -pg-url is specified but -target-db is not postgres. -pg-url will be ignored.")
	}

	if cfg.WebMode {
		return cfg, nil
	}

	if cfg.CompletionShell != "" {
		script, err := generateCompletionScript(cfg.CompletionShell)
		if err != nil {
			return nil, err
		}
		fmt.Print(script)
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
