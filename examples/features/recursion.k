// Recursive function call.

let factorial = (n) -> if n <= 1 { 1 } else { n * factorial(n - 1) }
let result = factorial(5)
result
