package interpreter

import "sync"

type Environment struct {
	mu    sync.RWMutex
	store map[string]Value
	outer *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]Value)}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func (e *Environment) Get(name string) (Value, bool) {
	e.mu.RLock()
	val, ok := e.store[name]
	e.mu.RUnlock()
	if ok {
		return val, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}

func (e *Environment) Define(name string, val Value) {
	e.mu.Lock()
	e.store[name] = val
	e.mu.Unlock()
}

func (e *Environment) Set(name string, val Value) bool {
	e.mu.Lock()
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		e.mu.Unlock()
		return true
	}
	e.mu.Unlock()
	if e.outer != nil {
		return e.outer.Set(name, val)
	}
	return false
}

func (e *Environment) Snapshot() map[string]Value {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]Value, len(e.store))
	for k, v := range e.store {
		out[k] = v
	}
	return out
}
