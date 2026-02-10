package interpreter

import (
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
)

func builtinEncodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "encodeJson expects 1 argument"}
	}
	value, err := encodeJSONValue(args[0])
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, &RuntimeError{Message: "encodeJson error: " + err.Error()}
	}
	return &String{Value: string(data)}, nil
}

func encodeJSONValue(value Value) (interface{}, error) {
	switch v := value.(type) {
	case *Null:
		return nil, nil
	case *Boolean:
		return v.Value, nil
	case *Integer:
		return v.Value, nil
	case *Float:
		if math.IsNaN(v.Value) || math.IsInf(v.Value, 0) {
			return nil, &RuntimeError{Message: "encodeJson cannot encode NaN or Inf"}
		}
		return v.Value, nil
	case *String:
		return v.Value, nil
	case *Char:
		return v.Value, nil
	case *Array:
		out := make([]interface{}, 0, len(v.Elements))
		for _, el := range v.Elements {
			enc, err := encodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, enc)
		}
		return out, nil
	case *Object:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *ModuleObject:
		if v.Env == nil {
			return map[string]interface{}{}, nil
		}
		out := make(map[string]interface{})
		for k, val := range v.Env.Snapshot() {
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = enc
		}
		return out, nil
	case *Map:
		out := make(map[string]interface{}, len(v.Pairs))
		for k, val := range v.Pairs {
			if k.Type != STRING && k.Type != CHAR {
				return nil, &RuntimeError{Message: "encodeJson only supports map with string keys"}
			}
			enc, err := encodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k.Value] = enc
		}
		return out, nil
	default:
		return nil, &RuntimeError{Message: "encodeJson does not support " + string(value.Type())}
	}
}

func builtinDecodeJSON(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "decodeJson expects 1 argument"}
	}
	str, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "decodeJson expects string"}
	}
	decoder := json.NewDecoder(strings.NewReader(str.Value))
	decoder.UseNumber()
	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, recoverableError("decodeJson", "decodeJson error: "+err.Error())
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, recoverableError("decodeJson", "decodeJson expects a single JSON value")
	}
	return decodeJSONValue(data)
}

func decodeJSONValue(value interface{}) (Value, error) {
	switch v := value.(type) {
	case nil:
		return NullValue, nil
	case bool:
		return &Boolean{Value: v}, nil
	case string:
		return &String{Value: v}, nil
	case json.Number:
		raw := v.String()
		if strings.ContainsAny(raw, ".eE") {
			f, err := strconv.ParseFloat(raw, 64)
			if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
				return nil, recoverableError("decodeJson", "decodeJson invalid float: "+raw)
			}
			return &Float{Value: f}, nil
		}
		i, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, recoverableError("decodeJson", "decodeJson integer overflow: "+raw)
		}
		return &Integer{Value: i}, nil
	case []interface{}:
		out := make([]Value, 0, len(v))
		for _, el := range v {
			val, err := decodeJSONValue(el)
			if err != nil {
				return nil, err
			}
			out = append(out, val)
		}
		return &Array{Elements: out}, nil
	case map[string]interface{}:
		out := make(map[string]Value, len(v))
		for k, val := range v {
			decoded, err := decodeJSONValue(val)
			if err != nil {
				return nil, err
			}
			out[k] = decoded
		}
		return &Object{Pairs: out}, nil
	default:
		return nil, recoverableError("decodeJson", "decodeJson unsupported value")
	}
}
