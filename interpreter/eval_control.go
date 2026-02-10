package interpreter

import "karl/ast"

func (e *Evaluator) evalRecoverExpression(node *ast.RecoverExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Target, env)
	if err == nil && sig == nil {
		return val, nil, nil
	}
	if sig != nil {
		return val, sig, err
	}
	switch err.(type) {
	case *RecoverableError, *RuntimeError:
		// recover block handles both recoverable and runtime errors.
	default:
		return nil, nil, err
	}
	fallbackEnv := NewEnclosedEnvironment(env)
	fallbackEnv.Define("error", errorValue(err))
	return e.Eval(node.Fallback, fallbackEnv)
}

// isTruthy determines if a value is truthy according to Karl's truthy/falsy rules
// Falsy values: null, false, 0, "", [], {}
// Everything else is truthy
func isTruthy(val Value) bool {
	switch v := val.(type) {
	case *Null:
		return false
	case *Boolean:
		return v.Value
	case *Integer:
		return v.Value != 0
	case *Float:
		return v.Value != 0.0
	case *String:
		return v.Value != ""
	case *Array:
		return len(v.Elements) > 0
	case *Object:
		return len(v.Pairs) > 0
	case *Map:
		return len(v.Pairs) > 0
	case *Set:
		return len(v.Elements) > 0
	default:
		// All other types are truthy (functions, tasks, channels, etc.)
		return true
	}
}

func (e *Evaluator) evalIfExpression(node *ast.IfExpression, env *Environment) (Value, *Signal, error) {
	cond, sig, err := e.Eval(node.Condition, env)
	if err != nil || sig != nil {
		return cond, sig, err
	}
	// Support truthy/falsy evaluation
	if isTruthy(cond) {
		return e.Eval(node.Consequence, env)
	}
	if node.Alternative != nil {
		return e.Eval(node.Alternative, env)
	}
	return UnitValue, nil, nil
}

func (e *Evaluator) evalBlockExpression(block *ast.BlockExpression, env *Environment) (Value, *Signal, error) {
	blockEnv := NewEnclosedEnvironment(env)
	var result Value = UnitValue
	for _, stmt := range block.Statements {
		val, sig, err := e.Eval(stmt, blockEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		result = val
	}
	return result, nil, nil
}

func (e *Evaluator) evalMatchExpression(node *ast.MatchExpression, env *Environment) (Value, *Signal, error) {
	value, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return value, sig, err
	}
	for _, arm := range node.Arms {
		armEnv := NewEnclosedEnvironment(env)
		ok, err := matchPattern(arm.Pattern, value, armEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		if arm.Guard != nil {
			guardVal, sig, err := e.Eval(arm.Guard, armEnv)
			if err != nil || sig != nil {
				return guardVal, sig, err
			}
			// Support truthy/falsy evaluation for match guards
			if !isTruthy(guardVal) {
				continue
			}
		}
		return e.Eval(arm.Body, armEnv)
	}
	return nil, nil, &RuntimeError{Message: "non-exhaustive match"}
}

func (e *Evaluator) evalForExpression(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	loopEnv := NewEnclosedEnvironment(env)
	for _, binding := range node.Bindings {
		val, sig, err := e.Eval(binding.Value, loopEnv)
		if err != nil || sig != nil {
			return val, sig, err
		}
		ok, err := bindPattern(binding.Pattern, val, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			return nil, nil, &RuntimeError{Message: "for binding pattern did not match"}
		}
	}

	for {
		condVal, sig, err := e.Eval(node.Condition, loopEnv)
		if err != nil || sig != nil {
			return condVal, sig, err
		}
		// Support truthy/falsy evaluation for loop condition
		if !isTruthy(condVal) {
			break
		}

		_, sig, err = e.Eval(node.Body, loopEnv)
		if err != nil {
			return nil, nil, err
		}
		if sig != nil {
			switch sig.Type {
			case SignalContinue:
				continue
			case SignalBreak:
				if sig.Value != nil {
					return sig.Value, nil, nil
				}
				// break with no value: evaluate then block
				return e.evalThen(node, loopEnv)
			default:
				return nil, sig, nil
			}
		}
	}

	return e.evalThen(node, loopEnv)
}

func (e *Evaluator) evalThen(node *ast.ForExpression, env *Environment) (Value, *Signal, error) {
	if node.Then == nil {
		return UnitValue, nil, nil
	}
	return e.Eval(node.Then, env)
}

func (e *Evaluator) evalBreakExpression(node *ast.BreakExpression, env *Environment) (Value, *Signal, error) {
	if node.Value == nil {
		return UnitValue, &Signal{Type: SignalBreak}, nil
	}
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	return val, &Signal{Type: SignalBreak, Value: val}, nil
}
