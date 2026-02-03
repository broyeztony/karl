// Truthy/falsy across common values and operators.

let show = (label, value) -> {
    let verdict = if value { "truthy" } else { "falsy" }
    log(label, "=>", verdict)
}

// Falsy values
show("null", null)
show("false", false)
show("0", 0)
show("0.0", 0.0)
show("empty string", "")
show("empty array", [])
show("empty object", {})
show("empty map", map())
show("empty set", set())

// Truthy values
show("true", true)
show("1", 1)
show("-1", -1)
show("1.5", 1.5)
show("\"hello\"", "hello")
show("non-empty array", [1])
show("non-empty object", { a: 1 })
show("non-empty map", map().set("a", 1))
show("non-empty set", set().add(1))

// Logical operators return Bool and use truthy/falsy
log("0 || 1 =>", 0 || 1)
log("0 && 1 =>", 0 && 1)
log("!\"\" =>", !"")
log("!42 =>", !42)

// For loop condition uses truthy/falsy
let countdown = for n with n = 3, acc = [] {
    acc += [n]
    n = n - 1
} then acc
countdown

// Match guards use truthy/falsy
let guarded = match 1 {
    case _ if "" -> "nope"
    case _ -> "ok"
}
guarded
