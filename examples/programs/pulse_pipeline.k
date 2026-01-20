let ch = rendezvous()

let ping = (ch) -> {
    ch.send(1)
    ch.done()
}

let pong = (ch) -> {
    let [v, done] = ch.recv()
    if done { 0 } else { v }
}

let t = & { ping(ch), pong(ch) }
wait t
