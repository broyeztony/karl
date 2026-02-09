// Join (& { ... }) is fail-fast:
// - if any child fails, the join task fails,
// - remaining work is canceled automatically.

let state = { hits: 0 }

let slowSideEffect = () -> {
    sleep(120)
    state.hits = state.hits + 1
    "slow-ok"
}

let boom = () -> fail("boom")

let join = & { boom(), slowSideEffect() }
let result = (wait join) ? {
    log("join failed:", error.kind, "-", error.message)
    "fallback"
}

// Give slowSideEffect time to run *if it wasn't canceled*.
sleep(200)

let output = { result, hits: state.hits }
output

