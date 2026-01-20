// String helpers return new values.

let text = " Hello Karl "
let trimmed = text.trim()
let lower = trimmed.toLower()
let upper = trimmed.toUpper()
let parts = trimmed.split("l")
let chars = trimmed.chars()
let has = trimmed.contains("ell")
let starts = trimmed.startsWith("He")
let ends = trimmed.endsWith("lo")
let replaced = trimmed.replace("l", "x")
let output = [trimmed, lower, upper, parts.length, chars.length, has, starts, ends, replaced]
output
