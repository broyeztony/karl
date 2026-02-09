// A task can fail with either a runtime error (e.g. missing property) or a recoverable error (fail()).
// The error is stored on the task and surfaces when you wait it (where you can recover with ?).

let unsafe = () -> {
    let obj = {}
    obj.missing // runtime error: missing property
}

let t1 = & unsafe()
let out1 = (wait t1) ? {
    log("caller recovered:", error.kind, "-", error.message)
    "fallback-from-caller"
}

// Recover inside the task: the task completes successfully (caller sees a normal value).
let safe = () -> {
    let obj = {}
    obj.missing ? {
        log("task recovered:", error.kind, "-", error.message)
        null
    }
}

let t2 = & safe()
let out2 = wait t2

// then() composes tasks: errors propagate through then until you recover.
let base = & (() -> 1)()
let chained = base.then(x -> x + 1).then(x -> x + 1)
let out3 = wait chained

let output = [out1, out2, out3]
output

