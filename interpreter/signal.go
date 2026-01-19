package interpreter

type SignalType int

const (
	SignalNone SignalType = iota
	SignalBreak
	SignalContinue
)

type Signal struct {
	Type  SignalType
	Value Value
}
