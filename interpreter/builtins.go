package interpreter

var builtins = map[string]*Builtin{}

func RegisterBuiltins() {
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
	builtins["readFile"] = &Builtin{Name: "readFile", Fn: builtinReadFile}
	builtins["writeFile"] = &Builtin{Name: "writeFile", Fn: builtinWriteFile}
	builtins["appendFile"] = &Builtin{Name: "appendFile", Fn: builtinAppendFile}
	builtins["deleteFile"] = &Builtin{Name: "deleteFile", Fn: builtinDeleteFile}
	builtins["exists"] = &Builtin{Name: "exists", Fn: builtinExists}
	builtins["listDir"] = &Builtin{Name: "listDir", Fn: builtinListDir}
	builtins["http"] = &Builtin{Name: "http", Fn: builtinHTTP}
	builtins["encodeJson"] = &Builtin{Name: "encodeJson", Fn: builtinEncodeJSON}
	builtins["decodeJson"] = &Builtin{Name: "decodeJson", Fn: builtinDecodeJSON}
	builtins["then"] = &Builtin{Name: "then", Fn: builtinThen}
	builtins["send"] = &Builtin{Name: "send", Fn: builtinSend}
	builtins["recv"] = &Builtin{Name: "recv", Fn: builtinRecv}
	builtins["done"] = &Builtin{Name: "done", Fn: builtinDone}
	builtins["split"] = &Builtin{Name: "split", Fn: builtinSplit}
	builtins["chars"] = &Builtin{Name: "chars", Fn: builtinChars}
	builtins["trim"] = &Builtin{Name: "trim", Fn: builtinTrim}
	builtins["toLower"] = &Builtin{Name: "toLower", Fn: builtinToLower}
	builtins["toUpper"] = &Builtin{Name: "toUpper", Fn: builtinToUpper}
	builtins["contains"] = &Builtin{Name: "contains", Fn: builtinContains}
	builtins["startsWith"] = &Builtin{Name: "startsWith", Fn: builtinStartsWith}
	builtins["endsWith"] = &Builtin{Name: "endsWith", Fn: builtinEndsWith}
	builtins["replace"] = &Builtin{Name: "replace", Fn: builtinReplace}
	builtins["map"] = &Builtin{Name: "map", Fn: builtinMap}
	builtins["get"] = &Builtin{Name: "get", Fn: builtinMapGet}
	builtins["set"] = &Builtin{Name: "set", Fn: builtinMapSet}
	builtins["add"] = &Builtin{Name: "add", Fn: builtinSetAdd}
	builtins["has"] = &Builtin{Name: "has", Fn: builtinMapHas}
	builtins["delete"] = &Builtin{Name: "delete", Fn: builtinMapDelete}
	builtins["keys"] = &Builtin{Name: "keys", Fn: builtinMapKeys}
	builtins["values"] = &Builtin{Name: "values", Fn: builtinMapValues}
	builtins["sort"] = &Builtin{Name: "sort", Fn: builtinSort}
	builtins["filter"] = &Builtin{Name: "filter", Fn: builtinFilter}
	builtins["reduce"] = &Builtin{Name: "reduce", Fn: builtinReduce}
	builtins["sum"] = &Builtin{Name: "sum", Fn: builtinSum}
	builtins["find"] = &Builtin{Name: "find", Fn: builtinFind}
	builtins["abs"] = &Builtin{Name: "abs", Fn: builtinAbs}
	builtins["sqrt"] = &Builtin{Name: "sqrt", Fn: builtinSqrt}
	builtins["pow"] = &Builtin{Name: "pow", Fn: builtinPow}
	builtins["sin"] = &Builtin{Name: "sin", Fn: builtinSin}
	builtins["cos"] = &Builtin{Name: "cos", Fn: builtinCos}
	builtins["tan"] = &Builtin{Name: "tan", Fn: builtinTan}
	builtins["floor"] = &Builtin{Name: "floor", Fn: builtinFloor}
	builtins["ceil"] = &Builtin{Name: "ceil", Fn: builtinCeil}
	builtins["min"] = &Builtin{Name: "min", Fn: builtinMin}
	builtins["max"] = &Builtin{Name: "max", Fn: builtinMax}
	builtins["clamp"] = &Builtin{Name: "clamp", Fn: builtinClamp}
}

func getBuiltin(name string) *Builtin {
	if b, ok := builtins[name]; ok {
		return b
	}
	return nil
}

func NewBaseEnvironment() *Environment {
	RegisterBuiltins()
	env := NewEnvironment()
	for name, builtin := range builtins {
		env.Define(name, builtin)
	}
	return env
}

func bindReceiver(fn BuiltinFunction, receiver Value) BuiltinFunction {
	return func(e *Evaluator, args []Value) (Value, error) {
		return fn(e, append([]Value{receiver}, args...))
	}
}

func recoverableError(kind string, msg string) *RecoverableError {
	return &RecoverableError{Kind: kind, Message: msg}
}
