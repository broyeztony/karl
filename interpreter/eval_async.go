package interpreter

import (
	"karl/ast"
	"sort"
)

func (e *Evaluator) handleAsyncError(task *Task, err error) {
	if err == nil {
		return
	}
	if exitErr, ok := err.(*ExitError); ok {
		exitProcess(exitErr.Message)
		return
	}
	if task == nil {
		return
	}
	task.complete(nil, err)
	if e.runtime == nil {
		return
	}
	if e.runtime.getTaskFailurePolicy() != TaskFailurePolicyFailFast {
		return
	}
	if task.internal || task.isObserved() {
		return
	}
	formatted := FormatRuntimeError(err, task.source, task.filename)
	e.runtime.setFatalTaskFailure(&UnhandledTaskError{Messages: []string{formatted}})
}

func (e *Evaluator) CheckUnhandledTaskFailures() error {
	if e.runtime == nil {
		return nil
	}
	tasks := e.runtime.snapshotTasks()
	msgs := []string{}
	for _, t := range tasks {
		if t == nil || t.internal {
			continue
		}
		if !t.isDone() || t.isObserved() {
			continue
		}
		err := t.getError()
		if err == nil {
			continue
		}
		if re, ok := err.(*RecoverableError); ok && re.Kind == "canceled" {
			continue
		}
		msgs = append(msgs, FormatRuntimeError(err, t.source, t.filename))
	}
	if len(msgs) == 0 {
		return nil
	}
	sort.Strings(msgs)
	return &UnhandledTaskError{Messages: msgs}
}

func (e *Evaluator) evalAwaitExpression(node *ast.AwaitExpression, env *Environment) (Value, *Signal, error) {
	val, sig, err := e.Eval(node.Value, env)
	if err != nil || sig != nil {
		return val, sig, err
	}
	task, ok := val.(*Task)
	if !ok {
		return nil, nil, &RuntimeError{Message: "wait expects task"}
	}
	var cancelCh <-chan struct{}
	if e.currentTask != nil {
		cancelCh = e.currentTask.cancelCh
	}
	return taskAwaitWithCancel(task, cancelCh, e.runtime)
}

func (e *Evaluator) evalRaceExpression(node *ast.RaceExpression, env *Environment) (Value, *Signal, error) {
	raceTask := e.newTask(e.currentTask, false)

	children := make([]*Task, 0, len(node.Tasks))
	for _, taskExpr := range node.Tasks {
		child, err := e.spawnTask(taskExpr, env, raceTask, true)
		if err != nil {
			return nil, nil, err
		}
		children = append(children, child)
	}

	go func() {
		type result struct {
			value Value
			sig   *Signal
			err   error
		}
		results := make(chan result, len(children))
		for _, child := range children {
			go func(t *Task) {
				val, sig, err := taskAwaitWithCancel(t, raceTask.cancelCh, e.runtime)
				results <- result{value: val, sig: sig, err: err}
			}(child)
		}

		select {
		case <-raceTask.cancelCh:
			// canceled by user or parent; Cancel() already completed the task.
			return
		case first := <-results:
			// Cancel losers. Cancellation is cooperative; losers will stop at the next yield point.
			raceTask.cancelChildren()
			if first.sig != nil {
				raceTask.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
				return
			}
			raceTask.complete(first.value, first.err)
			return
		}
	}()

	return raceTask, nil, nil
}

func (e *Evaluator) evalSpawnExpression(node *ast.SpawnExpression, env *Environment) (Value, *Signal, error) {
	if node.Task != nil {
		task, err := e.spawnTask(node.Task, env, e.currentTask, false)
		if err != nil {
			return nil, nil, err
		}
		return task, nil, nil
	}

	join := e.newTask(e.currentTask, false)

	children := make([]*Task, 0, len(node.Group))
	for _, expr := range node.Group {
		child, err := e.spawnTask(expr, env, join, true)
		if err != nil {
			return nil, nil, err
		}
		children = append(children, child)
	}

	go func() {
		type result struct {
			idx   int
			value Value
			sig   *Signal
			err   error
		}

		resultsCh := make(chan result, len(children))
		for i, child := range children {
			go func(idx int, t *Task) {
				val, sig, err := taskAwaitWithCancel(t, join.cancelCh, e.runtime)
				resultsCh <- result{idx: idx, value: val, sig: sig, err: err}
			}(i, child)
		}

		out := make([]Value, len(children))
		remaining := len(children)
		for remaining > 0 {
			select {
			case <-join.cancelCh:
				// canceled by user or parent; Cancel() already completed the task.
				return
			case r := <-resultsCh:
				if r.err != nil {
					// Fail fast: cancel remaining children and surface the error on the join task.
					join.cancelChildren()
					join.complete(nil, r.err)
					return
				}
				if r.sig != nil {
					join.cancelChildren()
					join.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
					return
				}
				out[r.idx] = r.value
				remaining--
			}
		}

		join.complete(&Array{Elements: out}, nil)
	}()

	return join, nil, nil
}

func (e *Evaluator) spawnTask(expr ast.Expression, env *Environment, parent *Task, internal bool) (*Task, error) {
	task := e.newTask(parent, internal)
	taskEval := e.cloneForTask(task)
	go func() {
		val, sig, err := taskEval.Eval(expr, env)
		if err != nil {
			taskEval.handleAsyncError(task, err)
			return
		}
		if sig != nil {
			task.complete(nil, &RuntimeError{Message: "break/continue outside loop"})
			return
		}
		task.complete(val, nil)
	}()
	return task, nil
}
