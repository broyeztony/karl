// import returns an object of bindings.

let util = import "examples/features/import_module.k"
let msg = util.greet("karl")
let output = [msg, util.answer]
output
