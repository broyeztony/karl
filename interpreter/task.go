package interpreter

import (
	"os"
)

func newTask() *Task {
	return &Task{
		ResultCh: make(chan taskResult, 1),
		cancelCh: make(chan struct{}),
	}
}

func (t *Task) complete(value Value, err error) {
	t.mu.Lock()
	if t.done {
		t.mu.Unlock()
		return
	}
	t.done = true
	t.result = value
	t.err = err
	t.mu.Unlock()

	t.ResultCh <- taskResult{value: value, err: err}
}

func taskAwait(t *Task) (Value, *Signal, error) {
	return taskAwaitWithCancel(t, nil, nil)
}

func taskAwaitWithCancel(t *Task, cancelCh <-chan struct{}, runtime *runtimeState) (Value, *Signal, error) {
	if t == nil {
		return nil, nil, &RuntimeError{Message: "wait expects task"}
	}

	t.markObserved()

	t.mu.Lock()
	if t.done {
		res := t.result
		err := t.err
		t.mu.Unlock()
		if err != nil {
			return nil, nil, err
		}
		return res, nil, nil
	}
	t.mu.Unlock()

	var out taskResult
	fatalCh := runtime.fatalSignal()

	if cancelCh == nil && fatalCh == nil {
		out = <-t.ResultCh
	} else {
		select {
		case out = <-t.ResultCh:
		case <-cancelCh:
			return nil, nil, canceledError()
		case <-fatalCh:
			if err := runtime.getFatalTaskFailure(); err != nil {
				return nil, nil, err
			}
			return nil, nil, &RuntimeError{Message: "runtime terminated"}
		}
	}

	t.mu.Lock()
	t.done = true
	t.result = out.value
	t.err = out.err
	t.mu.Unlock()

	if out.err != nil {
		return nil, nil, out.err
	}
	return out.value, nil, nil
}

func (t *Task) markObserved() {
	t.mu.Lock()
	t.observed = true
	t.mu.Unlock()
}

func (t *Task) isObserved() bool {
	t.mu.Lock()
	observed := t.observed
	t.mu.Unlock()
	return observed
}

func (t *Task) isDone() bool {
	t.mu.Lock()
	done := t.done
	t.mu.Unlock()
	return done
}

func (t *Task) getError() error {
	t.mu.Lock()
	err := t.err
	t.mu.Unlock()
	return err
}

func (t *Task) addChild(child *Task) {
	if t == nil || child == nil {
		return
	}
	t.mu.Lock()
	t.children = append(t.children, child)
	t.mu.Unlock()
}

func (t *Task) cancelChildren() {
	t.mu.Lock()
	children := append([]*Task(nil), t.children...)
	t.mu.Unlock()
	for _, child := range children {
		child.Cancel()
	}
}

func (t *Task) Cancel() {
	if t == nil {
		return
	}
	t.cancelOnce.Do(func() { close(t.cancelCh) })
	t.cancelChildren()
	t.complete(nil, canceledError())
}

func (t *Task) canceled() bool {
	if t == nil {
		return false
	}
	select {
	case <-t.cancelCh:
		return true
	default:
		return false
	}
}

func exitProcess(msg string) {
	if msg != "" {
		_, _ = os.Stderr.WriteString(msg + "\n")
	}
	os.Exit(1)
}

func canceledError() *RecoverableError {
	return &RecoverableError{Kind: "canceled", Message: "task canceled"}
}
