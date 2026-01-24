let retry = (max) -> {
    let last = "no attempts"
    for i < max with i = 0 {
        if i == 2 { break "ok" }
        i++
    } then last
}

retry(3)
