package dialect

import (
	"strings"
	"testing"
)

func TestValidateOracleIdentifier(t *testing.T) {
	valid := []string{
		"USERS",
		"MY_TABLE_1",
		"SYS_C00123",
		"TABLE_1",
		"_PRIVATE",
		"A",
		strings.Repeat("A", 128), // 최대 128자
	}
	for _, name := range valid {
		if err := ValidateOracleIdentifier(name); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", name, err)
		}
	}

	invalid := []struct {
		name   string
		reason string
	}{
		{"", "empty string"},
		{"1TABLE", "starts with digit"},
		{`"DROP TABLE`, "contains double quote"},
		{"; SELECT *", "contains semicolon"},
		{"MY TABLE", "contains space"},
		{"TABLE-1", "contains hyphen"},
		{strings.Repeat("A", 129), "exceeds 128 chars"},
	}
	for _, tc := range invalid {
		if err := ValidateOracleIdentifier(tc.name); err == nil {
			t.Errorf("expected %q (%s) to be invalid, but got no error", tc.name, tc.reason)
		}
	}
}

func TestQuoteOracleIdentifier(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"USERS", `"USERS"`},
		{"MY_TABLE", `"MY_TABLE"`},
		{`MY"TABLE`, `"MY""TABLE"`},
		{`A""B`, `"A""""B"`},
	}
	for _, tc := range cases {
		got := QuoteOracleIdentifier(tc.input)
		if got != tc.expected {
			t.Errorf("QuoteOracleIdentifier(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
