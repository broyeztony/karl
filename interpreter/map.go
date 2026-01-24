package interpreter

import "strconv"

func mapKeyForValue(val Value) (MapKey, error) {
	switch v := val.(type) {
	case *String:
		return MapKey{Type: STRING, Value: v.Value}, nil
	case *Char:
		return MapKey{Type: CHAR, Value: v.Value}, nil
	case *Integer:
		return MapKey{Type: INTEGER, Value: strconv.FormatInt(v.Value, 10)}, nil
	case *Boolean:
		if v.Value {
			return MapKey{Type: BOOLEAN, Value: "true"}, nil
		}
		return MapKey{Type: BOOLEAN, Value: "false"}, nil
	default:
		return MapKey{}, &RuntimeError{Message: "map keys must be string, char, int, or bool"}
	}
}

func mapKeyToValue(key MapKey) Value {
	switch key.Type {
	case STRING:
		return &String{Value: key.Value}
	case CHAR:
		return &Char{Value: key.Value}
	case INTEGER:
		val, err := strconv.ParseInt(key.Value, 10, 64)
		if err != nil {
			return &String{Value: key.Value}
		}
		return &Integer{Value: val}
	case BOOLEAN:
		return &Boolean{Value: key.Value == "true"}
	default:
		return &String{Value: key.Value}
	}
}

func formatMapKey(key MapKey) string {
	return mapKeyToValue(key).Inspect()
}
