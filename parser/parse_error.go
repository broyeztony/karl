package parser

import (
	"fmt"
	"strings"

	"karl/token"
)

type ParseError struct {
	Message string
	Token   token.Token
}

func FormatParseErrors(errs []ParseError, source string, filename string) string {
	if len(errs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(errs))
	for _, err := range errs {
		parts = append(parts, formatParseError(err, source, filename))
	}
	return strings.Join(parts, "\n")
}

func formatParseError(err ParseError, source string, filename string) string {
	if err.Token.Line == 0 || source == "" {
		return "parse error: " + err.Message
	}
	lines := strings.Split(source, "\n")
	line := err.Token.Line
	col := err.Token.Column
	if line < 1 || line > len(lines) {
		return "parse error: " + err.Message
	}
	lineText := strings.TrimRight(lines[line-1], "\r")
	if col < 1 {
		col = 1
	}
	if col > len(lineText)+1 {
		col = len(lineText) + 1
	}
	caret := strings.Repeat(" ", col-1) + "^"
	location := fmt.Sprintf("%d:%d", line, err.Token.Column)
	if filename != "" {
		location = fmt.Sprintf("%s:%s", filename, location)
	}
	return fmt.Sprintf(
		"parse error: %s\n  at %s\n  %d | %s\n    | %s",
		err.Message,
		location,
		line,
		lineText,
		caret,
	)
}
