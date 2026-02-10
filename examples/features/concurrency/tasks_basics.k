// Tasks are futures: & spawns work; wait awaits the value; then chains a continuation.
// Join: & { ... } runs calls concurrently and returns results in *input order*.

let work = (label, ms, value) -> {
    log(label, "start")
    sleep(ms)
    log(label, "done")
    value
}

let t = & work("A", 60, 1)
let t2 = t.then(x -> x + 1)
let v = wait t2

let all = wait & {
    work("B", 40, 2),
    work("C", 20, 3)
}

// v is 2; all is [2, 3] (input order), even if C finishes before B.
let output = [v, all]
output
