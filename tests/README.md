# Tests

This folder holds unit tests for the Karl lexer, parser, and interpreter.

## What lives here

- `lexer_test.go` — tokenization correctness
- `parser_test.go` — core parse coverage
- `basic_expressions_test.go` — expression shapes by kind
- `object_disambiguation_test.go` — block vs object literal rules
- `parser_errors_test.go` — expected parser failures
- `interpreter_eval_test.go` — runtime evaluation behavior
- `examples_test.go` — parses all example programs

## Adding a test

1) Add a focused unit test next to the closest file above.
2) Keep inputs small and assert AST shape or evaluated value.
3) If a feature needs a full-file fixture, add a `.k` file under `examples/`
   and ensure `examples_test.go` keeps parsing it.

## Running

```
go test ./...
```

If the Go build cache is locked, run:

```
GOCACHE=$(pwd)/.go-cache go test ./...
```
