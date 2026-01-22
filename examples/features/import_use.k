// import returns a factory; calling it returns bindings.

let makeUtil = import "examples/features/import_module.k"
let util = makeUtil()
let msg = util.greet("karl")
let output = [msg, util.answer]
output
