package interpreter

import (
	"fmt"
	"strings"
	"time"
)

func registerRuntimeCoreBuiltins() {
	builtins["exit"] = &Builtin{Name: "exit", Fn: builtinExit}
	builtins["fail"] = &Builtin{Name: "fail", Fn: builtinFail}
	builtins["rendezvous"] = &Builtin{Name: "rendezvous", Fn: builtinChannel}
	builtins["channel"] = &Builtin{Name: "channel", Fn: builtinChannel}
	builtins["buffered"] = &Builtin{Name: "buffered", Fn: builtinBufferedChannel}
	builtins["sleep"] = &Builtin{Name: "sleep", Fn: builtinSleep}
	builtins["log"] = &Builtin{Name: "log", Fn: builtinLog}
	builtins["str"] = &Builtin{Name: "str", Fn: builtinStr}
}

func runtimeFatalSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.runtime == nil {
		return nil
	}
	return e.runtime.fatalSignal()
}

func runtimeCancelSignal(e *Evaluator) <-chan struct{} {
	if e == nil || e.currentTask == nil {
		return nil
	}
	return e.currentTask.cancelCh
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
	cancelCh := runtimeCancelSignal(e)
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
