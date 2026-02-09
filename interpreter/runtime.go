package interpreter

import (
	"fmt"
	"sync"
)

const (
	TaskFailurePolicyFailFast = "fail-fast"
	TaskFailurePolicyDefer    = "defer"
)

func validTaskFailurePolicy(policy string) bool {
	return policy == TaskFailurePolicyFailFast || policy == TaskFailurePolicyDefer
}

// runtimeState is shared across Evaluator instances (main file, imported modules,
// spawned tasks). It lets us inspect task outcomes at the end of a run.
type runtimeState struct {
	mu                sync.Mutex
	tasks             map[*Task]struct{}
	taskFailurePolicy string
	fatalTaskFailure  error
}

func newRuntimeState() *runtimeState {
	return &runtimeState{
		tasks:             make(map[*Task]struct{}),
		taskFailurePolicy: TaskFailurePolicyFailFast,
	}
}

func (r *runtimeState) registerTask(t *Task) {
	if r == nil || t == nil {
		return
	}
	r.mu.Lock()
	r.tasks[t] = struct{}{}
	r.mu.Unlock()
}

func (r *runtimeState) snapshotTasks() []*Task {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	out := make([]*Task, 0, len(r.tasks))
	for t := range r.tasks {
		out = append(out, t)
	}
	r.mu.Unlock()
	return out
}

func (r *runtimeState) setTaskFailurePolicy(policy string) error {
	if r == nil {
		return nil
	}
	if !validTaskFailurePolicy(policy) {
		return fmt.Errorf("invalid task failure policy: %s", policy)
	}
	r.mu.Lock()
	r.taskFailurePolicy = policy
	r.mu.Unlock()
	return nil
}

func (r *runtimeState) getTaskFailurePolicy() string {
	if r == nil {
		return TaskFailurePolicyFailFast
	}
	r.mu.Lock()
	policy := r.taskFailurePolicy
	r.mu.Unlock()
	if policy == "" {
		return TaskFailurePolicyFailFast
	}
	return policy
}

func (r *runtimeState) setFatalTaskFailure(err error) {
	if r == nil || err == nil {
		return
	}
	r.mu.Lock()
	if r.fatalTaskFailure == nil {
		r.fatalTaskFailure = err
	}
	r.mu.Unlock()
}

func (r *runtimeState) getFatalTaskFailure() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	err := r.fatalTaskFailure
	r.mu.Unlock()
	return err
}
