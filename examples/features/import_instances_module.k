// Module with per-instance state.

let count = 0
let inc = () -> {
    count += 1
    count
}
let get = () -> count
