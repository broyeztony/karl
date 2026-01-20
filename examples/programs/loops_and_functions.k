let sum = for i <= 10 with i = 1, total = 0 {
    total += i
    i++
} then total

let firstEven = for true with i = 0 {
    i++
    if i % 2 == 1 { continue }
    break i
} then 0

let squares = for i < 4 with i = 0, acc = [] {
    acc += [i * i]
    i++
} then acc;

[sum, firstEven, squares]
