# Changelog

All notable changes to Karl are documented here.

## [v0.3.2] - 2026-02-03

Highlights:
- Added buffered channels via `buffered(size)`.
- Added `channel()` as an alias for `rendezvous()`.
- CLI no longer prints `()` when the program result is `Unit`.

Docs & examples:
- Documented `channel()` alias in specs and examples.
- Added a buffered channels feature example.

Developer:
- Ignore `.vscode/` in `.gitignore`.

## [v0.3.1] - 2026-02-03

Highlights:
- Truthy/falsy semantics across conditionals and logical operators (Python-style; empty map/set are falsy).
- New truthy/falsy feature examples.
- Added a full workflow-engine example suite (contrib).

Interpreter:
- `if`, `for`, and match guards now use truthy/falsy rules.
- `!`, `&&`, and `||` operate on truthy/falsy and return `Bool`.
- Empty `map()` and `set()` are falsy; non-empty are truthy.

Tests:
- Added unit tests covering truthy/falsy evaluation, short-circuiting, loop conditions, and guards.

Examples:
- Added `examples/features/truthy_falsy.k` and `examples/features/truthy_falsy_comprehensive.k`.
- Added `examples/contrib/workflow/*` (DAG engine, pipelines, timers, file watcher, quickstart).

Docs:
- Updated `SPECS/interpreter.md` with truthy/falsy rules.
- Updated `examples/README.md` with new examples.

Developer:
- Ignore `*.karl-new` and `karl` artifacts in `.gitignore`.
