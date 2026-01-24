let nums = 1..5
let users = [
    { name: "a", age: 20 },
    { name: "b", age: 15 },
]

let adults = from u in users
    where u.age >= 18
    select { name: u.name, age: u.age }

let stats = { count: adults.length, last: nums[nums.length - 1] }
stats
