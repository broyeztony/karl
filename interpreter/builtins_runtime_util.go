package interpreter

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

func registerRuntimeUtilityBuiltins() {
	builtins["rand"] = &Builtin{Name: "rand", Fn: builtinRand}
	builtins["randInt"] = &Builtin{Name: "randInt", Fn: builtinRandInt}
	builtins["randFloat"] = &Builtin{Name: "randFloat", Fn: builtinRandFloat}
	builtins["parseInt"] = &Builtin{Name: "parseInt", Fn: builtinParseInt}
	builtins["now"] = &Builtin{Name: "now", Fn: builtinNow}
}

func builtinRand(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "rand expects no arguments"}
	}
	return &Integer{Value: rand.Int63()}, nil
}

func builtinRandInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randInt expects min and max"}
	}
	min, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer min"}
	}
	max, ok := args[1].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "randInt expects integer max"}
	}
	if max.Value < min.Value {
		return nil, &RuntimeError{Message: "randInt expects min <= max"}
	}
	if max.Value == min.Value {
		return &Integer{Value: min.Value}, nil
	}
	diff := max.Value - min.Value
	if diff < 0 || diff == math.MaxInt64 {
		return nil, &RuntimeError{Message: "randInt range too large"}
	}
	n := rand.Int63n(diff+1) + min.Value
	return &Integer{Value: n}, nil
}

func builtinRandFloat(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "randFloat expects min and max"}
	}
	min, _, ok := numberArg(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric min"}
	}
	max, _, ok := numberArg(args[1])
	if !ok {
		return nil, &RuntimeError{Message: "randFloat expects numeric max"}
	}
	if max < min {
		return nil, &RuntimeError{Message: "randFloat expects min <= max"}
	}
	return &Float{Value: min + rand.Float64()*(max-min)}, nil
}

func builtinParseInt(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "parseInt expects 1 argument"}
	}
	s, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "parseInt expects string"}
	}
	n, err := strconv.ParseInt(s.Value, 10, 64)
	if err != nil {
		return nil, &RuntimeError{Message: "invalid integer: " + s.Value}
	}
	return &Integer{Value: n}, nil
}

func builtinNow(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "now expects no arguments"}
	}
	return &Integer{Value: time.Now().UnixNano() / int64(time.Millisecond)}, nil
}
