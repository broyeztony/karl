package interpreter

import (
	"encoding/json"
	"testing"

	"karl/shape"
)

func TestApplyShapeAliasAndDrop(t *testing.T) {
	shapeText := `HttpResponse : object
    + headers : object
        + acceptEncoding(Accept-Encoding) : string
        - userAgent(User Agent) : string
    + status : int
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	value := &Object{Pairs: map[string]Value{
		"headers": &Object{Pairs: map[string]Value{
			"Accept-Encoding": &String{Value: "gzip"},
			"User Agent":      &String{Value: "Karl"},
			"Extra":           &String{Value: "drop"},
		}},
		"status": &Integer{Value: 200},
		"extra":  &String{Value: "drop"},
	}}
	out, err := applyShape(value, sh.Type)
	if err != nil {
		t.Fatalf("applyShape error: %v", err)
	}
	obj, ok := out.(*Object)
	if !ok {
		t.Fatalf("expected object, got %T", out)
	}
	if _, ok := obj.Pairs["extra"]; ok {
		t.Fatalf("unexpected extra field")
	}
	headers, ok := obj.Pairs["headers"].(*Object)
	if !ok {
		t.Fatalf("expected headers object")
	}
	if _, ok := headers.Pairs["acceptEncoding"]; !ok {
		t.Fatalf("missing acceptEncoding")
	}
	if _, ok := headers.Pairs["userAgent"]; !ok {
		t.Fatalf("missing userAgent")
	}

	enc, err := encodeJSONValue(out)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}
	raw, err := json.Marshal(enc)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}
	var decoded map[string]interface{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	headersOut, ok := decoded["headers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected headers map in json output")
	}
	if _, ok := headersOut["Accept-Encoding"]; !ok {
		t.Fatalf("expected alias key Accept-Encoding")
	}
	if _, ok := headersOut["User Agent"]; !ok {
		t.Fatalf("expected alias key User Agent")
	}
}

func TestApplyShapeMissingRequired(t *testing.T) {
	shapeText := `User : object
    + name : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	value := &Object{Pairs: map[string]Value{}}
	_, err = applyShape(value, sh.Type)
	if err == nil {
		t.Fatalf("expected error for missing required field")
	}
	if _, ok := err.(*RecoverableError); !ok {
		t.Fatalf("expected recoverable error, got %T", err)
	}
}

func TestShapeCallEquivalent(t *testing.T) {
	shapeText := `User : object
    + name : string
`
	sh, err := shape.Parse(shapeText)
	if err != nil {
		t.Fatalf("shape parse error: %v", err)
	}
	eval := NewEvaluator()
	val, sig, err := eval.applyFunction(&ShapeValue{Name: sh.Name, Shape: sh.Type}, []Value{
		&Object{Pairs: map[string]Value{"name": &String{Value: "Karl"}}},
	})
	if err != nil {
		t.Fatalf("shape call error: %v", err)
	}
	if sig != nil {
		t.Fatalf("unexpected signal")
	}
	obj, ok := val.(*Object)
	if !ok || obj.Pairs["name"].(*String).Value != "Karl" {
		t.Fatalf("unexpected shaped value")
	}
}
