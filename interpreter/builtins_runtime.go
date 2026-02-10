package interpreter

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func registerRuntimeBuiltins() {
	builtins["exit"] = &Builtin{Name: "exit", Fn: builtinExit}
	builtins["fail"] = &Builtin{Name: "fail", Fn: builtinFail}
	builtins["rendezvous"] = &Builtin{Name: "rendezvous", Fn: builtinChannel}
	builtins["channel"] = &Builtin{Name: "channel", Fn: builtinChannel}
	builtins["buffered"] = &Builtin{Name: "buffered", Fn: builtinBufferedChannel}
	builtins["sleep"] = &Builtin{Name: "sleep", Fn: builtinSleep}
	builtins["log"] = &Builtin{Name: "log", Fn: builtinLog}
	builtins["str"] = &Builtin{Name: "str", Fn: builtinStr}
	builtins["rand"] = &Builtin{Name: "rand", Fn: builtinRand}
	builtins["randInt"] = &Builtin{Name: "randInt", Fn: builtinRandInt}
	builtins["randFloat"] = &Builtin{Name: "randFloat", Fn: builtinRandFloat}
	builtins["parseInt"] = &Builtin{Name: "parseInt", Fn: builtinParseInt}
	builtins["now"] = &Builtin{Name: "now", Fn: builtinNow}
}

func runtimeFatalSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.fatalSignal()
}

func runtimeFatalError(e *Evaluator) error {
	if e != nil && e.runtime != nil {
		if err := e.runtime.getFatalTaskFailure(); err != nil {
			return err
		}
	}
	return &RuntimeError{Message: "runtime terminated"}
}

func builtinExit(_ *Evaluator, args []Value) (Value, error) {
	msg := ""
	if len(args) > 0 {
		msg = args[0].Inspect()
	}
	exitProcess(msg)
	return nil, &ExitError{Message: msg}
}

func builtinFail(_ *Evaluator, args []Value) (Value, error) {
	if len(args) > 1 {
		return nil, &RuntimeError{Message: "fail expects 0 or 1 argument"}
	}
	msg := ""
	if len(args) == 1 {
		s, ok := args[0].(*String)
		if !ok {
			return nil, &RuntimeError{Message: "fail expects string message"}
		}
		msg = s.Value
	}
	return nil, recoverableError("fail", msg)
}

func builtinChannel(_ *Evaluator, _ []Value) (Value, error) {
	return &Channel{Ch: make(chan Value)}, nil
}

func builtinBufferedChannel(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "buffered expects 1 argument (buffer size)"}
	}
	size, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "buffered expects integer buffer size"}
	}
	if size.Value < 0 {
		return nil, &RuntimeError{Message: "buffered expects non-negative buffer size"}
	}
	if size.Value > 1000000 {
		return nil, &RuntimeError{Message: "buffered buffer size too large (max 1000000)"}
	}
	return &Channel{Ch: make(chan Value, size.Value)}, nil
}

func builtinSleep(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "sleep expects 1 argument"}
	}
	ms, ok := args[0].(*Integer)
	if !ok {
		return nil, &RuntimeError{Message: "sleep expects integer milliseconds"}
	}

	d := time.Duration(ms.Value) * time.Millisecond
	if d <= 0 {
		return UnitValue, nil
	}
	fatalCh := runtimeFatalSignal(e)
	cancelCh := (<-chan struct{})(nil)
	if e != nil && e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	if cancelCh == nil && fatalCh == nil {
		time.Sleep(d)
		return UnitValue, nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-timer.C:
		return UnitValue, nil
	case <-cancelCh:
		return nil, canceledError()
	case <-fatalCh:
		return nil, runtimeFatalError(e)
	}
}

func builtinLog(_ *Evaluator, args []Value) (Value, error) {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatLogValue(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return UnitValue, nil
}

func builtinStr(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "str expects 1 argument"}
	}
	return &String{Value: formatLogValue(args[0])}, nil
}

func formatLogValue(val Value) string {
	switch v := val.(type) {
	case *String:
		return v.Value
	case *Char:
		return v.Value
	case *Null:
		return "null"
	case *Unit:
		return "()"
	default:
		return val.Inspect()
	}
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
