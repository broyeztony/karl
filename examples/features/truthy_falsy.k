// Truthy/falsy basics: empty values are falsy; non-empty are truthy.

let config = {
    retryPolicy: { maxAttempts: 3 },
    storage: null,
}

if config.retryPolicy {
    log("✅ Has retry policy")
} else {
    log("❌ No retry policy")
}

if config.storage {
    log("✅ Has storage")
} else {
    log("❌ No storage")
}

let tasks = []
let status = if tasks { "has tasks" } else { "no tasks" }
status
