// task.cancel() requests cancellation (cooperative).
// - race cancels losers automatically
// - join cancels remaining work on first error
// - you can also cancel manually

let state = { done: false }

let long = () -> {
    sleep(200)
    state.done = true
    "done"
}

let t = & long()
sleep(20)
t.cancel()

let out = (wait t) ? { "canceled (" + error.kind + ")" }

// Wait a bit to confirm long() did not complete its side-effect.
sleep(250)

// Cancel a whole group task (it cancels its children too).
let state2 = { hits: 0 }
let slow1 = () -> { sleep(200); state2.hits = state2.hits + 1; 1 }
let slow2 = () -> { sleep(200); state2.hits = state2.hits + 1; 2 }
let join = & { slow1(), slow2() }
sleep(20)
join.cancel()

let joined = (wait join) ? { error.kind }
sleep(250)

let output = { out, done: state.done, joined, hits: state2.hits }
output

