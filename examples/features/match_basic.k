// Match with range patterns and wildcard.

let value = 3
let label = match value {
    case 0 -> "zero"
    case 1..4 -> "small"
    case _ -> "other"
}
label
