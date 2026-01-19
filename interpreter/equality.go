package interpreter

func StrictEqual(left, right Value) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *Integer:
		return l.Value == right.(*Integer).Value
	case *Float:
		return l.Value == right.(*Float).Value
	case *Boolean:
		return l.Value == right.(*Boolean).Value
	case *String:
		return l.Value == right.(*String).Value
	case *Char:
		return l.Value == right.(*Char).Value
	case *Null, *Unit:
		return true
	case *Map:
		return left == right
	default:
		// Arrays, objects, functions, tasks, channels: identity equality.
		return left == right
	}
}

func Equivalent(left, right Value) bool {
	if left == nil || right == nil {
		return left == right
	}
	if left.Type() != right.Type() {
		return false
	}

	switch l := left.(type) {
	case *Integer:
		return l.Value == right.(*Integer).Value
	case *Float:
		return l.Value == right.(*Float).Value
	case *Boolean:
		return l.Value == right.(*Boolean).Value
	case *String:
		return l.Value == right.(*String).Value
	case *Char:
		return l.Value == right.(*Char).Value
	case *Null, *Unit:
		return true
	case *Array:
		r := right.(*Array)
		if len(l.Elements) != len(r.Elements) {
			return false
		}
		for i := range l.Elements {
			if !Equivalent(l.Elements[i], r.Elements[i]) {
				return false
			}
		}
		return true
	case *Object:
		r := right.(*Object)
		if len(l.Pairs) != len(r.Pairs) {
			return false
		}
		for k, v := range l.Pairs {
			ov, ok := r.Pairs[k]
			if !ok {
				return false
			}
			if !Equivalent(v, ov) {
				return false
			}
		}
		return true
	case *Map:
		r := right.(*Map)
		if len(l.Pairs) != len(r.Pairs) {
			return false
		}
		for k, v := range l.Pairs {
			ov, ok := r.Pairs[k]
			if !ok {
				return false
			}
			if !Equivalent(v, ov) {
				return false
			}
		}
		return true
	default:
		return left == right
	}
}
