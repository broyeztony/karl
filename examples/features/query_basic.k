// Query expressions filter and project collections.

let users = [
    { name: "a", age: 20 },
    { name: "b", age: 15 },
    { name: "c", age: 30 },
]

let adults = from u in users
    where u.age >= 18
    orderby u.name
    select u.name

adults
