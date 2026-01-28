// Destructuring patterns in let bindings (trailing commas allowed).

let point = Point { x: 10, y: 20, }
let { x, y } = point
let { x: x2, y: y2, } = point

let nums = [1, 2, 3, 4]
let pairNums = [1, 2]
let [first, second] = pairNums
let [head, ...tail, ] = nums

let pair = [7, 9]
let (a, b) = pair
let (c, d, ) = pair

let output = [x, y, x2, y2, first, second, head, tail, a, b, c, d]
output
