package interpreter

import (
	"fmt"
	"karl/ast"
	"unicode/utf8"
)

func (e *Evaluator) evalCallExpression(node *ast.CallExpression, env *Environment) (Value, *Signal, error) {
	function, sig, err := e.Eval(node.Function, env)
	if err != nil || sig != nil {
		return function, sig, err
	}

	args := make([]Value, 0, len(node.Arguments))
	hasPlaceholder := false
	for _, arg := range node.Arguments {
		if _, ok := arg.(*ast.Placeholder); ok {
			args = append(args, nil)
			hasPlaceholder = true
			continue
		}
		val, sig, err := e.Eval(arg, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		args = append(args, val)
	}

	if hasPlaceholder {
		return &Partial{Target: function, Args: args}, nil, nil
	}
	return e.applyFunction(function, args)
}

func (e *Evaluator) evalMemberExpression(node *ast.MemberExpression, env *Environment) (Value, *Signal, error) {
	object, sig, err := e.Eval(node.Object, env)
	if err != nil || sig != nil {
		return object, sig, err
	}

	switch obj := object.(type) {
	case *Object:
		val, ok := obj.Pairs[node.Property.Value]
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *ModuleObject:
		if obj.Env == nil {
			return nil, nil, &RuntimeError{Message: "member access on invalid module object"}
		}
		val, ok := obj.Env.GetLocal(node.Property.Value)
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + node.Property.Value}
		}
		return val, nil, nil
	case *Array:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.arrayMethod(obj, node.Property.Value)
	case *String:
		if node.Property.Value == "length" {
			return &Integer{Value: int64(utf8.RuneCountInString(obj.Value))}, nil, nil
		}
		return e.stringMethod(obj, node.Property.Value)
	case *Map:
		return e.mapMethod(obj, node.Property.Value)
	case *Set:
		if node.Property.Value == "size" {
			return &Integer{Value: int64(len(obj.Elements))}, nil, nil
		}
		return e.setMethod(obj, node.Property.Value)
	case *Channel:
		return e.channelMethod(obj, node.Property.Value)
	case *Task:
		return e.taskMethod(obj, node.Property.Value)
	default:
		if object == nil {
			return nil, nil, &RuntimeError{Message: "member access on non-object (got <nil>)"}
		}
		return nil, nil, &RuntimeError{Message: fmt.Sprintf("member access on non-object (%s.%s)", object.Type(), node.Property.Value)}
	}
}

func (e *Evaluator) evalIndexExpression(node *ast.IndexExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	indexVal, sig, err := e.Eval(node.Index, env)
	if err != nil || sig != nil {
		return indexVal, sig, err
	}

	switch indexed := left.(type) {
	case *Array:
		idx, ok := indexVal.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "index must be integer"}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(indexed.Elements) {
			return nil, nil, &RuntimeError{Message: "index out of bounds"}
		}
		return indexed.Elements[i], nil, nil
	case *Object:
		key, ok := objectIndexKey(indexVal)
		if !ok {
			return nil, nil, &RuntimeError{Message: "object index must be string or char"}
		}
		val, ok := indexed.Pairs[key]
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + key}
		}
		return val, nil, nil
	case *ModuleObject:
		if indexed.Env == nil {
			return nil, nil, &RuntimeError{Message: "indexing requires array or object"}
		}
		key, ok := objectIndexKey(indexVal)
		if !ok {
			return nil, nil, &RuntimeError{Message: "object index must be string or char"}
		}
		val, ok := indexed.Env.GetLocal(key)
		if !ok {
			return nil, nil, &RuntimeError{Message: "missing property: " + key}
		}
		return val, nil, nil
	default:
		return nil, nil, &RuntimeError{Message: "indexing requires array or object"}
	}
}

func (e *Evaluator) evalSliceExpression(node *ast.SliceExpression, env *Environment) (Value, *Signal, error) {
	left, sig, err := e.Eval(node.Left, env)
	if err != nil || sig != nil {
		return left, sig, err
	}
	arr, ok := left.(*Array)
	if !ok {
		return nil, nil, &RuntimeError{Message: "slice requires array"}
	}

	start := 0
	end := len(arr.Elements)

	if node.Start != nil {
		val, sig, err := e.Eval(node.Start, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "slice start must be integer"}
		}
		start = normalizeIndex(int(idx.Value), len(arr.Elements))
	}

	if node.End != nil {
		val, sig, err := e.Eval(node.End, env)
		if err != nil || sig != nil {
			return val, sig, err
		}
		idx, ok := val.(*Integer)
		if !ok {
			return nil, nil, &RuntimeError{Message: "slice end must be integer"}
		}
		end = normalizeIndex(int(idx.Value), len(arr.Elements))
	}

	if start < 0 || start > len(arr.Elements) || end < 0 || end > len(arr.Elements) {
		return nil, nil, &RuntimeError{Message: "slice bounds out of range"}
	}
	if start >= end {
		return &Array{Elements: []Value{}}, nil, nil
	}
	out := append([]Value{}, arr.Elements[start:end]...)
	return &Array{Elements: out}, nil, nil
}

func normalizeIndex(idx int, length int) int {
	if idx < 0 {
		return length + idx
	}
	return idx
}
