# Concurrency in Karl (Tasks, Failures, Recovery)

Karl concurrency is built around **Tasks** (futures): `&` starts work concurrently and returns a handle.
You **observe** a task with `wait`.

Core operators:
- `& call()` spawns a task and returns a task handle.
- `& { call1(), call2(), ... }` spawns multiple tasks and returns a *single* task handle; `wait` yields results in input order.
- `| { call1(), call2(), ... }` races multiple tasks and returns a *single* task handle; `wait` yields the first completion.
- `task.then(fn)` attaches a continuation and returns a new task handle.

Errors:
- A task completes with either a **value** or an **error**.
- Errors are stored on the task and surface on `wait`.
- Use `? { ... }` (or `? fallbackExpr`) around `wait ...` to recover.
- Default policy is `fail-fast`: detached unobserved task failures fail the run quickly.
- You can opt into deferred reporting with `karl run ... --task-failure-policy=defer`.

Cancellation:
- `task.cancel()` requests cancellation for a task (and its children).
- `| { ... }` cancels losers automatically.
- `& { ... }` cancels remaining work on first error (fail-fast).
- Cancellation is cooperative; it takes effect while waiting/blocked (e.g. `wait`, `sleep`, `send`, `recv`, `http`).

Suggested reading order:
1) `tasks_basics.k`
2) `then_and_errors.k`
3) `join_fail_fast.k`
4) `race_timeout.k`
5) `cancellation.k`
6) `channels_and_cancel.k`
7) `timeout_wrapper.k`
8) `unhandled_failures.k`
