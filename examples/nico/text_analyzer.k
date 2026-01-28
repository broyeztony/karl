// Text Analyzer - Advanced String Processing
// Analyzes text for word frequency, palindromes, and patterns

// Sample text to analyze
let text = "A man a plan a canal Panama racecar level deed"

// Normalize text: lowercase and split into words
let words = text.toLower().split(" ")

// Count word frequencies
let word_frequencies = (words) -> {
    for i < words.length with i = 0, freq_map = map() {
        let word = words[i]
        if freq_map.has(word) {
            let current = freq_map.get(word)
            freq_map.set(word, current + 1)
        } else {
            freq_map.set(word, 1)
        }
        i++
    } then freq_map
}

// Check if a string is a palindrome
let is_palindrome = (word) -> {
    let chars = word.chars()
    let len = chars.length
    
    for i < len / 2 with i = 0, is_pal = true {
        if chars[i] != chars[len - 1 - i] {
            is_pal = false
            break is_pal
        }
        i++
    } then is_pal
}

// Find all palindromes in word list
let find_palindromes = (words) -> {
    words.filter(word -> {
        word.length > 2 && is_palindrome(word)
    })
}

// Get word frequency map
let frequencies = word_frequencies(words)

// Find palindromes
let palindromes = find_palindromes(words)

// Find longest word
let longest = for i < words.length with i = 0, longest_word = "" {
    if words[i].length > longest_word.length {
        longest_word = words[i]
    }
    i++
} then longest_word

// Calculate total characters
let total_chars = for i < words.length with i = 0, total = 0 {
    total += words[i].length
    i++
} then total

// Find words that appear more than once
let repeated_words = for i < frequencies.keys().length with i = 0, repeated = [] {
    let word_key = frequencies.keys()[i]
    let count = frequencies.get(word_key)
    
    if count > 1 {
        repeated += [{ word: word_key, count: count, }]
    }
    
    i++
} then repeated

// Build statistics object
let stats = {
    total_words: words.length,
    unique_words: frequencies.keys().length,
    palindromes: palindromes,
    palindrome_count: palindromes.length,
    longest_word: longest,
    total_characters: total_chars,
}

log("📝 Text Analysis Results")
log("Original text:", text)
log("\n📊 Statistics:")
log("  Total words:", stats.total_words)
log("  Unique words:", stats.unique_words)
log("  Total characters:", stats.total_characters)
log("  Longest word:", stats.longest_word)
log("\n🔄 Palindromes found:", palindromes)
log("\n🔁 Repeated words:", repeated_words)

match stats.palindrome_count {
    case 0 -> log("\n❌ No palindromes found")
    case 1 -> log("\n✅ Found 1 palindrome!")
    case _ -> log("\n🎉 Found multiple palindromes!")
}

{
    statistics: stats,
    word_frequencies: repeated_words,
    all_words: words,
}
