// for is an expression; break/continue control flow.

let stopped = for i < 5 with i = 0, total = 0 {
    if i == 3 { break total }
    total += i
    i++
} then total

let sumOdds = for i < 6 with i = 0, acc = 0 {
    i++
    if i % 2 == 0 { continue }
    acc += i
} then acc

let output = { stopped, sumOdds, }
output
