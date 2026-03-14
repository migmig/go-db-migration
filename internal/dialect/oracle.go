package dialect

import (
	"fmt"
	"regexp"
	"strings"
)

// oracleIdentifierPattern은 Oracle 식별자 규칙을 따르는 패턴이다.
// 알파벳/밑줄로 시작, 영숫자/밑줄/$/#만 허용, 최대 128자.
var oracleIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$#]{0,127}$`)

// ValidateOracleIdentifier는 문자열이 유효한 Oracle 식별자인지 검증한다.
// 유효하지 않으면 에러를 반환한다.
func ValidateOracleIdentifier(name string) error {
	if name == "" {
		return fmt.Errorf("oracle identifier must not be empty")
	}
	if !oracleIdentifierPattern.MatchString(name) {
		return fmt.Errorf("invalid Oracle identifier: %q (must start with letter/underscore, contain only alphanumeric/_/$/#, max 128 chars)", name)
	}
	return nil
}

// QuoteOracleIdentifier는 Oracle 식별자를 큰따옴표로 감싸 이스케이프한다.
// 내부 큰따옴표는 두 번 반복(SQL 표준)하여 이스케이프한다.
func QuoteOracleIdentifier(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}
