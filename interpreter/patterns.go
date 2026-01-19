package interpreter

import (
	"fmt"
	"karl/ast"
)

func matchPattern(pattern ast.Pattern, value Value, env *Environment) (bool, error) {
	switch p := pattern.(type) {
	case *ast.WildcardPattern:
		return true, nil
	case *ast.Identifier:
		env.Define(p.Value, value)
		return true, nil
	case *ast.IntegerLiteral:
		v, ok := value.(*Integer)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.FloatLiteral:
		v, ok := value.(*Float)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.StringLiteral:
		v, ok := value.(*String)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.CharLiteral:
		v, ok := value.(*Char)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.BooleanLiteral:
		v, ok := value.(*Boolean)
		if !ok {
			return false, nil
		}
		return v.Value == p.Value, nil
	case *ast.NullLiteral:
		_, ok := value.(*Null)
		return ok, nil
	case *ast.RangePattern:
		return matchRangePattern(p, value)
	case *ast.ObjectPattern:
		return matchObjectPattern(p, value, env)
	case *ast.ArrayPattern:
		return matchArrayPattern(p, value, env)
	case *ast.TuplePattern:
		return matchTuplePattern(p, value, env)
	case *ast.CallPattern:
		return matchCallPattern(p, value, env)
	default:
		return false, fmt.Errorf("unsupported pattern type: %T", pattern)
	}
}

func matchRangePattern(p *ast.RangePattern, value Value) (bool, error) {
	start, err := patternLiteralValue(p.Start)
	if err != nil {
		return false, err
	}
	end, err := patternLiteralValue(p.End)
	if err != nil {
		return false, err
	}

	switch v := value.(type) {
	case *Integer:
		s, sok := start.(*Integer)
		e, eok := end.(*Integer)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	case *Float:
		s, sok := start.(*Float)
		e, eok := end.(*Float)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	case *Char:
		s, sok := start.(*Char)
		e, eok := end.(*Char)
		if !sok || !eok {
			return false, nil
		}
		return v.Value >= s.Value && v.Value <= e.Value, nil
	default:
		return false, nil
	}
}

func matchObjectPattern(p *ast.ObjectPattern, value Value, env *Environment) (bool, error) {
	obj, ok := value.(*Object)
	if !ok {
		return false, nil
	}
	for _, entry := range p.Entries {
		val, ok := obj.Pairs[entry.Key]
		if !ok {
			return false, nil
		}
		ok, err := matchPattern(entry.Pattern, val, env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchArrayPattern(p *ast.ArrayPattern, value Value, env *Environment) (bool, error) {
	arr, ok := value.(*Array)
	if !ok {
		return false, nil
	}
	if p.Rest == nil && len(arr.Elements) != len(p.Elements) {
		return false, nil
	}
	if p.Rest != nil && len(arr.Elements) < len(p.Elements) {
		return false, nil
	}
	for i, el := range p.Elements {
		ok, err := matchPattern(el, arr.Elements[i], env)
		if err != nil || !ok {
			return ok, err
		}
	}
	if p.Rest != nil {
		rest := &Array{Elements: append([]Value{}, arr.Elements[len(p.Elements):]...)}
		ok, err := matchPattern(p.Rest, rest, env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchTuplePattern(p *ast.TuplePattern, value Value, env *Environment) (bool, error) {
	arr, ok := value.(*Array)
	if !ok {
		return false, nil
	}
	if len(arr.Elements) != len(p.Elements) {
		return false, nil
	}
	for i, el := range p.Elements {
		ok, err := matchPattern(el, arr.Elements[i], env)
		if err != nil || !ok {
			return ok, err
		}
	}
	return true, nil
}

func matchCallPattern(p *ast.CallPattern, value Value, env *Environment) (bool, error) {
	return false, fmt.Errorf("call patterns are not supported")
}

func patternLiteralValue(pattern ast.Pattern) (Value, error) {
	switch p := pattern.(type) {
	case *ast.IntegerLiteral:
		return &Integer{Value: p.Value}, nil
	case *ast.FloatLiteral:
		return &Float{Value: p.Value}, nil
	case *ast.StringLiteral:
		return &String{Value: p.Value}, nil
	case *ast.CharLiteral:
		return &Char{Value: p.Value}, nil
	default:
		return nil, fmt.Errorf("range patterns require literal endpoints")
	}
}
