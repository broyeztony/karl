# Karl TODO

## Current priorities

- Keep tests green as syntax/runtime changes land (`gotest`).
- Parser: consider treating newlines as statement boundaries to reduce adjacency ambiguity.
- Extend test coverage when new syntax is added (parser + interpreter + examples).
- Brainstorm objects versus maps versus mutability versus shapes
- Recover block that run for any situation where the runtime throws an expection?
- Real-world readiness gaps:
  - Additional codecs (csv/yaml/query/headers) + codec options
  - Error model consistency (recoverable vs fatal across indexing/member access/etc.)
  - Byte/encoding utilities (base64, url encode/decode, utf-8 concerns)
  - Date/time + duration parsing/formatting
  - Env/config access
  - CLI args/flags ergonomics
  - Packaging/conventions for multi-file projects
  - Minimal testing tools or assert builtins
  - Deterministic/ordered map option
  - REPL / fast dev loop
  - Server apps: HTTP listener / request handling primitives

## Known review points

- Disambiguation rules (block vs object) must stay consistent with examples and tests ✅
- Import/factory behavior and live exports should remain explicit in specs.
