package interpreter

import (
	"karl/ast"
	"sort"
	"unicode/utf8"
)

type queryRow struct {
	item Value
	key  Value
}

func (e *Evaluator) evalArrayLiteral(node *ast.ArrayLiteral, env *Environment) (Value, *Signal, error) {
	elements := make([]Value, 0, len(node.Elements))
	for _, el := range node.Elements {
		val, sig, err := e.Eval(el, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		elements = append(elements, val)
	}
	return &Array{Elements: elements}, nil, nil
}

func (e *Evaluator) evalObjectLiteral(node *ast.ObjectLiteral, env *Environment) (Value, *Signal, error) {
	obj := &Object{Pairs: make(map[string]Value)}
	for _, entry := range node.Entries {
		if entry.Spread {
			val, sig, err := e.Eval(entry.Value, env)
			if err != nil || sig != nil {
				return val, sig, err
			}
			pairs, ok := objectPairs(val)
			if !ok {
				return nil, nil, &RuntimeError{Message: "object spread requires object"}
			}
			for k, v := range pairs {
				obj.Pairs[k] = v
			}
			continue
		}

		val, sig, err := e.Eval(entry.Value, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		obj.Pairs[entry.Key] = val
	}
	return obj, nil, nil
}

func (e *Evaluator) evalStructInitExpression(node *ast.StructInitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.evalObjectLiteral(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, nil, nil
}

func (e *Evaluator) evalRangeExpression(node *ast.RangeExpression, env *Environment) (Value, *Signal, error) {
	startVal, sig, err := e.Eval(node.Start, env)
	if err != nil || sig != nil {
		return startVal, sig, err
	}
	endVal, sig, err := e.Eval(node.End, env)
	if err != nil || sig != nil {
		return endVal, sig, err
	}
	if _, ok := startVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}
	if _, ok := endVal.(*Float); ok {
		return nil, nil, &RuntimeError{Message: "float ranges are not allowed"}
	}

	stepVal := Value(&Integer{Value: 1})
	if node.Step != nil {
		stepVal, sig, err = e.Eval(node.Step, env)
		if err != nil || sig != nil {
			return stepVal, sig, err
		}
	}

	return buildRange(startVal, endVal, stepVal)
}

func buildRange(start, end, step Value) (Value, *Signal, error) {
	switch s := start.(type) {
	case *Integer:
		e, ok := end.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "integer range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildIntRange(s.Value, e.Value, st.Value), nil, nil
	case *Float:
		e, ok := end.(*Float)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		var stepVal float64
		switch st := step.(type) {
		case *Float:
			stepVal = st.Value
		case *Integer:
			stepVal = float64(st.Value)
		default:
			return nil, nil, &RuntimeError{Message: "float range step must be number"}
		}
		if stepVal == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildFloatRange(s.Value, e.Value, stepVal), nil, nil
	case *Char:
		e, ok := end.(*Char)
		if !ok {
			return nil, nil, &RuntimeError{Message: "range endpoints must match types"}
		}
		st, ok := step.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "char range step must be integer"}
		}
		if st.Value == 0 {
			return nil, nil, &RuntimeError{Message: "range step cannot be zero"}
		}
		return buildCharRange(s.Value, e.Value, st.Value), nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "unsupported range endpoint type"}
	}
}

func buildIntRange(start, end, step int64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i <= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	} else {
		for i := start; i >= end; i += step {
			out = append(out, &Integer{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildFloatRange(start, end, step float64) Value {
	out := []Value{}
	if step > 0 {
		for i := start; i < end; i += step {
			out = append(out, &Float{Value: i})
		}
	} else {
		for i := start; i > end; i += step {
			out = append(out, &Float{Value: i})
		}
	}
	return &Array{Elements: out}
}

func buildCharRange(start, end string, step int64) Value {
	sr, _ := utf8.DecodeRuneInString(start)
	er, _ := utf8.DecodeRuneInString(end)
	out := []Value{}
	if step > 0 {
		for r := sr; r <= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	} else {
		for r := sr; r >= er; r += rune(step) {
			out = append(out, &Char{Value: string(r)})
		}
	}
	return &Array{Elements: out}
}

func (e *Evaluator) evalQueryExpression(node *ast.QueryExpression, env *Environment) (Value, *Signal, error) {
	sourceVal, sig, err := e.Eval(node.Source, env)
	if err != nil || sig != nil {
		return sourceVal, sig, err
	}
	source, ok := sourceVal.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "query source must be array"}
	}

	rows := []queryRow{}

	for _, item := range source.Elements {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, item)

		keep := true
		for _, whereExpr := range node.Where {
			val, sig, err := e.Eval(whereExpr, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			b, ok := val.(*Boolean)
			if !ok {
				return nil, nil, &RuntimeError{Message: "query where must be bool"}
			}
			if !b.Value {
				keep = false
				break
			}
		}

		if !keep {
			continue
		}

		var key Value
		if node.OrderBy != nil {
			val, sig, err := e.Eval(node.OrderBy, rowEnv)
			if err != nil || sig != nil {
				return val, sig, err
			}
			key = val
		}
		rows = append(rows, queryRow{item: item, key: key})
	}

	if node.OrderBy != nil {
		if err := sortRows(rows); err != nil {
			return nil, nil, err
		}
	}

	results := []Value{}
	for _, r := range rows {
		rowEnv := NewEnclosedEnvironment(env)
		rowEnv.Define(node.Var.Value, r.item)
		val, sig, err := e.Eval(node.Select, rowEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		results = append(results, val)
	}
	return &Array{Elements: results}, nil, nil
}

func sortRows(rows []queryRow) error {
	if len(rows) == 0 {
		return nil
	}
	for _, row := range rows {
		if row.key == nil {
			return &RuntimeError{Message: "orderby requires comparable key"}
		}
	}

	switch rows[0].key.(type) {
	case *Integer, *Float, *String:
	default:
		return &RuntimeError{Message: "orderby key must be int, float, or string"}
	}
	firstType := rows[0].key.Type()
	for _, row := range rows {
		if row.key.Type() != firstType {
			return &RuntimeError{Message: "orderby keys must be the same type"}
		}
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return compareForSort(rows[i].key, rows[j].key)
	})
	return nil
}

func compareForSort(left, right Value) bool {
	switch l := left.(type) {
	case *Integer:
		return l.Value < right.(*Integer).Value
	case *Float:
		return l.Value < right.(*Float).Value
	case *String:
		return l.Value < right.(*String).Value
	default:
		return false
	}
}
