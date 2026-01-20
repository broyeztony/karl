// Closure captures an outer binding.

let makeAdder = (n) -> x -> n + x
let add10 = makeAdder(10)
let result = add10(5)
result
