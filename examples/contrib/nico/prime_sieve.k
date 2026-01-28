// Sieve of Eratosthenes - Prime Number Generator
// A classic algorithm to find all prime numbers up to a given limit

// Check if a number is prime (simple method for small numbers)
let is_prime = (n) -> {
    if n < 2 {
        false
    } else {
        for i < n with i = 2, prime = true {
            if i * i > n {
                break prime
            }
            if n % i == 0 {
                prime = false
                break prime
            }
            i++
        } then prime
    }
}

// Generate all primes up to max using trial division
let find_primes = (max) -> {
    for num < max with num = 2, primes = [] {
        if is_prime(num) {
            primes += [num]
        }
        num++
    } then primes
}

// Generate primes up to 100
let primes = find_primes(100)

// Categorize primes by range
let categorize = (primes) -> {
    let small = primes.filter(p -> p < 20)
    let medium = primes.filter(p -> p >= 20 && p < 50)
    let large = primes.filter(p -> p >= 50)
    
    {
        small: small,
        medium: medium,
        large: large,
        count: {
            small: small.length,
            medium: medium.length,
            large: large.length,
        },
    }
}

// Find twin primes (pairs where difference is 2)
let find_twin_primes = (primes) -> {
    for i < primes.length - 1 with i = 0, twins = [] {
        let current = primes[i]
        let next = primes[i + 1]
        
        if next - current == 2 {
            twins += [{ p1: current, p2: next, }]
        }
        
        i++
    } then twins
}

let categories = categorize(primes)
let twins = find_twin_primes(primes)

log("🔢 Prime Number Explorer")
log("All primes under 100:", primes)
log("Total count:", primes.length)
log("\nCategories:", categories)
log("\nTwin Primes:", twins)

{
    all_primes: primes,
    total: primes.length,
    categories: categories,
    twin_primes: twins,
    largest: primes[primes.length - 1],
}
