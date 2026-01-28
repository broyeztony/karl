// Fibonacci Sequence Generator
// Generates fibonacci numbers using multiple approaches:
// 1. Recursive function with memoization pattern
// 2. Iterative loop with state accumulation
// 3. List mapping and filtering

// Simple recursive fibonacci
let fib_recursive = (n) -> {
    if n <= 1 {
        n
    } else {
        fib_recursive(n - 1) + fib_recursive(n - 2)
    }
}

// Iterative fibonacci using for loop
let fib_iterative = (n) -> {
    if n <= 1 {
        n
    } else {
        for i < n with i = 2, a = 0, b = 1, result = 1 {
            result = a + b
            a = b
            b = result
            i++
        } then result
    }
}

// Generate first N fibonacci numbers as a list
let fib_sequence = (count) -> {
    for i < count with i = 0, sequence = [] {
        sequence += [fib_iterative(i)]
        i++
    } then sequence
}

// Generate fibonacci sequence
let fibs = fib_sequence(15)

// Find all fibonacci numbers less than 200
let small_fibs = fibs.filter(x -> x < 200)

// Sum of all even fibonacci numbers in the sequence
let even_sum = fibs.filter(x -> x % 2 == 0).sum()

// Create beautiful output object
let output = {
    first_15: fibs,
    under_200: small_fibs,
    even_sum: even_sum,
    recursive_10th: fib_recursive(10),
    iterative_10th: fib_iterative(10),
}

log("🔢 Fibonacci Sequence Explorer")
log("First 15 numbers:", fibs)
log("Numbers under 200:", small_fibs)
log("Sum of even numbers:", even_sum)

output
