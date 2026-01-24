let ch = rendezvous()

let producer = (ch) -> {
    ch.send(1)
    ch.send(2)
    ch.done()
}

let consumer = (ch) -> {
    for true with res = ch.recv(), acc = [] {
        let [v, done] = res
        if done { break acc }
        acc += [v]
        res = ch.recv()
    } then acc
}

let tasks = & { producer(ch), consumer(ch) }
let results = wait tasks

let fast = () -> "fast"
let slow = () -> { sleep(1); "slow" }
let first = wait | { slow(), fast() };

[results[1], first]