package interpreter

import (
	"fmt"
	"karl/shape"
)

func applyShape(value Value, t *shape.Type) (Value, error) {
	if t == nil {
		return value, nil
	}
	switch t.Kind {
	case shape.KindAny:
		return value, nil
	case shape.KindString:
		if _, ok := value.(*String); ok {
			return value, nil
		}
		return nil, recoverableError("shape", "expected string")
	case shape.KindInt:
		if _, ok := value.(*Integer); ok {
			return value, nil
		}
		return nil, recoverableError("shape", "expected int")
	case shape.KindFloat:
		if _, ok := value.(*Float); ok {
			return value, nil
		}
		return nil, recoverableError("shape", "expected float")
	case shape.KindBool:
		if _, ok := value.(*Boolean); ok {
			return value, nil
		}
		return nil, recoverableError("shape", "expected bool")
	case shape.KindNull:
		if _, ok := value.(*Null); ok {
			return value, nil
		}
		return nil, recoverableError("shape", "expected null")
	case shape.KindArray:
		arr, ok := value.(*Array)
		if !ok {
			return nil, recoverableError("shape", "expected array")
		}
		out := make([]Value, 0, len(arr.Elements))
		for _, el := range arr.Elements {
			shaped, err := applyShape(el, t.Elem)
			if err != nil {
				return nil, err
			}
			out = append(out, shaped)
		}
		return &Array{Elements: out}, nil
	case shape.KindObject:
		pairs, ok := shapeObjectPairs(value)
		if !ok {
			return nil, recoverableError("shape", "expected object")
		}
		out := make(map[string]Value, len(t.Fields))
		for _, field := range t.Fields {
			val, found := lookupShapeField(pairs, field)
			if !found {
				if field.Required {
					return nil, recoverableError("shape", fmt.Sprintf("missing required field: %s", field.Name))
				}
				out[field.Name] = NullValue
				continue
			}
			shaped, err := applyShape(val, field.Type)
			if err != nil {
				return nil, err
			}
			out[field.Name] = shaped
		}
		return &Object{Pairs: out, Shape: t}, nil
	case shape.KindUnion:
		return nil, recoverableError("shape", "union types are not supported yet")
	default:
		return nil, recoverableError("shape", "unsupported shape")
	}
}

type shapePairs struct {
	object map[string]Value
	mapp   *Map
}

func shapeObjectPairs(value Value) (*shapePairs, bool) {
	switch v := value.(type) {
	case *Object:
		return &shapePairs{object: v.Pairs}, true
	case *ModuleObject:
		if v.Env == nil {
			return nil, false
		}
		return &shapePairs{object: v.Env.Snapshot()}, true
	case *Map:
		return &shapePairs{mapp: v}, true
	default:
		return nil, false
	}
}

func lookupShapeField(pairs *shapePairs, field shape.Field) (Value, bool) {
	if pairs == nil {
		return nil, false
	}
	if pairs.object != nil {
		if val, ok := pairs.object[field.Name]; ok {
			return val, true
		}
		if field.Alias != "" {
			if val, ok := pairs.object[field.Alias]; ok {
				return val, true
			}
		}
		return nil, false
	}
	if pairs.mapp != nil {
		if val, ok := pairs.mapp.Pairs[MapKey{Type: STRING, Value: field.Name}]; ok {
			return val, true
		}
		if field.Alias != "" {
			if val, ok := pairs.mapp.Pairs[MapKey{Type: STRING, Value: field.Alias}]; ok {
				return val, true
			}
		}
		return nil, false
	}
	return nil, false
}
