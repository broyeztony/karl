// Shapes: strict record objects with alias mapping for external keys.

let HttpResponse = import "examples/features/shapes/HttpResponse.shape"

let raw = "{ \"stats\": 200, \"headers\": { \"Accept-Encoding\": \"gzip\", \"User Agent\": \"Karl\" } }"
let parsed = decodeJson(raw) as HttpResponse ? {

    log("Shaping error", error)
    { status: 0, headers: { acceptEncoding: "identity", }, }
}

log("parsed:", parsed)

parsed.headers.acceptEncoding = "deflate"

let out = encodeJson(parsed);

log(out)
