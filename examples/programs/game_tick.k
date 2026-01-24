let ticks = for i < 3 with i = 0, acc = [] {
    acc += [i]
    i++
} then acc

let sum = for i < ticks.length with i = 0, total = 0 {
    total += ticks[i]
    i++
} then total

sum
