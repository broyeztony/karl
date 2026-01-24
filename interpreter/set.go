package interpreter

import "strconv"

func setKeyForValue(val Value) (MapKey, error) {
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
		return MapKey{}, &RuntimeError{Message: "set values must be string, char, int, or bool"}
	}
}
