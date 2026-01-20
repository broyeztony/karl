// Functions are first-class values.

let add = (a, b) -> a + b
let inc = x -> x + 1
let value = inc(add(2, 3))
value
