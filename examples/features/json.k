// JSON in Karl: decode, query, shape, and re-encode.

// Raw JSON from an external source.
let raw = "{ \"status\": 200, \"headers\": { \"Accept-Encoding\": \"gzip\", \"User Agent\": \"Karl\" }, \"request\": { \"User-Agent\": \"Karl/1.0\" }, \"items\": [ { \"id\": 1 }, { \"id\": 2 } ] }"

// 1) Decode JSON into plain Karl objects/arrays.
let body = decodeJson(raw) ? { exit("invalid json") }

// 2) Access keys that are not valid identifiers via jsonPath.
let ua = jsonPath(body, "request['User-Agent']")
let firstId = jsonPath(body, "items[0].id")
log("ua:", ua)
log("first id:", firstId)

// 3) Optional shaping: map external keys to internal identifiers.
let HttpResponse = import "examples/features/shapes/HttpResponse.shape"
let shaped = body as HttpResponse

log("acceptEncoding:", shaped.headers.acceptEncoding)
log("userAgent:", shaped.headers.userAgent)

// 4) Modify and re-encode using aliases declared in the shape.
shaped.headers.acceptEncoding = "br"
let out = encodeJson(shaped)
log(out)
