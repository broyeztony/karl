// Unhandled task failures:
// - Default mode is fail-fast: detached unobserved task failures fail the run quickly.
// - In defer mode (`karl run ... --task-failure-policy=defer`), they are reported at program end.
// - So either (1) wait it (and recover), or (2) make the task recover internally.

let boom = () -> {
    let obj = {}
    obj.missing
}

// GOOD: observe it (wait) and decide how to handle the failure.
let t = & boom()
let v1 = (wait t) ? { "fallback" }

// ALSO GOOD: recover inside the task, so it never fails.
let safeBoom = () -> boom() ? {
    log("ignored task error:", error.message)
    null
}

let detached = & safeBoom()
// We intentionally don't wait `detached` (it can't fail).
sleep(10)

// BAD (example only - do not uncomment):
// & boom()
// sleep(10) // give it time to fail

let output = v1
output
