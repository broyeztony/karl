# Nico's Karl Examples

Three cool programs demonstrating different features of the Karl programming language.

## Programs

### 1. fibonacci_sequence.k
**Fibonacci Sequence Generator**

Demonstrates:
- Recursive functions
- Iterative loops with state management (`for` with `with` clause)
- List generation and manipulation
- Higher-order functions (`.filter()`, `.sum()`)
- Multiple implementation approaches to solve the same problem

Features:
- Generates the first 15 Fibonacci numbers
- Filters numbers under 200
- Calculates sum of even Fibonacci numbers
- Compares recursive vs iterative approaches

Run with:
```bash
cat examples/nico/fibonacci_sequence.k | ./karl run -
```

### 2. prime_sieve.k
**Prime Number Generator (Sieve of Eratosthenes)**

Demonstrates:
- Nested loop control flow
- Early breaking with `break`
- Complex conditionals with `if/else`
- Data categorization and filtering
- Array indexing and manipulation
- Object literal creation with shorthand syntax

Features:
- Finds all prime numbers under 100
- Categorizes primes by size (small/medium/large)
- Identifies twin primes (pairs with difference of 2)
- Comprehensive statistical analysis

Run with:
```bash
cat examples/nico/prime_sieve.k | ./karl run -
```

### 3. text_analyzer.k
**Text Analyzer - String Processing & Pattern Detection**

Demonstrates:
- String manipulation (`.toLower()`, `.split()`, `.chars()`)
- Map operations (`.set()`, `.get()`, `.has()`, `.keys()`)
- Palindrome detection algorithm
- Word frequency counting
- Pattern matching with `match`
- Complex data structure building

Features:
- Detects palindromes in text
- Counts word frequencies
- Finds repeated words
- Calculates text statistics
- Uses pattern matching for result reporting

Run with:
```bash
cat examples/nico/text_analyzer.k | ./karl run -
```

## Key Karl Features Demonstrated

### Control Flow
- `for` loops with `with` clause for loop-local state
- `break` and `continue` statements
- `then` clause for loop result expressions
- `if/else` expressions
- `match` pattern matching with ranges and wildcards

### Functions
- Arrow function syntax `() ->`
- Closures and higher-order functions
- Recursive functions

### Data Structures
- Arrays/Lists with methods: `.filter()`, `.map()`, `.sum()`, `.length`
- Maps with `.set()`, `.get()`, `.has()`, `.keys()`
- Object literals with shorthand and spread syntax

### String Processing
- String methods: `.toLower()`, `.split()`, `.chars()`
- Array indexing for character access

---

**Author:** Nico  
**Language:** Karl  
**Created:** 2026-01-28
