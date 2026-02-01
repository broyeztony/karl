# Proposal: Server Apps in Karl

This document sketches two directions for enabling server applications in Karl.
Both are **drafts** for reasoning and not implemented yet.

---

## Direction A — Low-level primitives (`listen`, `accept`, `serve`)

Expose a small set of runtime primitives and let users assemble servers with
loops and concurrency. This keeps the surface area minimal and composable.

### Proposed primitives

- `listen(addr)` -> Listener
- `accept(listener)` -> Connection (blocks)
- `serve(conn, handler)` -> Unit
  - Parses HTTP from `conn`, calls `handler(request)`
  - Writes response back to `conn`

### Example

```karl
let addr = "127.0.0.1:8080"
let server = listen(addr)

for true with _ = null {
    let conn = accept(server)
    & serve(conn, request -> {
        match request {
            case { method: "GET", path: "/health" } -> {
                { status: 200, headers: map(), body: "ok" }
            }
            case _ -> {
                { status: 404, headers: map(), body: "not found" }
            }
        }
    })
}
```

### Pros

- Minimal primitives
- Composable with Karl’s concurrency model (`&`, `wait`)
- Lets advanced users build custom protocols or middleware

### Cons

- More boilerplate
- Easier to write unsafe/inefficient servers

---

## Direction B — High-level `httpServer`

Provide a single high-level builtin that starts an HTTP server from a config
object. This is ergonomic and safer for most users.

### Proposed API

```
httpServer({ addr, handler, ...options }) -> Server
```

### Example

```karl
let server = httpServer({
    addr: "127.0.0.1:8080",
    handler: req -> {
        match req {
            case { method: "GET", path: "/health" } -> {
                { status: 200, headers: map(), body: "ok" }
            }
            case { method: "POST", path: "/echo", body } -> {
                { status: 200, headers: map(), body }
            }
            case _ -> {
                { status: 404, headers: map(), body: "not found" }
            }
        }
    }
})

wait server
```

### Pros

- Simple, low boilerplate
- Easier to standardize behavior (timeouts, header handling, keep-alive)
- Better entry point for beginners

### Cons

- Less flexible for protocol or middleware customization

---

## Open questions

- Should the request/response shapes be formalized (e.g., using shapes)?
- Should `handler` be recoverable by default (`? { ... }`)?
- How should streaming bodies be modeled (string vs bytes vs channel)?
- Do we want a middleware pattern or router built-in?
- How do we expose TLS configuration?
