// Cancellation also applies to blocked operations (recv/send/sleep/wait).

// Cancel a blocked recv().
let ch = channel()
let receiver = () -> {
    log("receiver: waiting for a message...")
    let [v, done] = ch.recv()
    if done { "closed" } else { v }
}

let tr = & receiver()
sleep(20)
tr.cancel()
let recvOut = (wait tr) ? { error.kind }

// Cancel a blocked send() (rendezvous has no receiver).
let ch2 = channel()
let sender = () -> { ch2.send("hi"); "sent" }

let ts = & sender()
sleep(20)
ts.cancel()
let sendOut = (wait ts) ? { error.kind }

let output = { recvOut, sendOut, }
output
