package tests

import (
	"karl/lexer"
	"karl/parser"
	"testing"
)

func TestParserErrors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "float_range_not_allowed",
			input: "1.0..2.0",
		},
		{
			name:  "recover_requires_call",
			input: "1 ? { 2 }",
		},
		{
			name:  "spawn_requires_call",
			input: "& foo",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p := parser.New(lexer.New(tc.input))
			_ = p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatalf("expected parse errors")
			}
		})
	}
}
