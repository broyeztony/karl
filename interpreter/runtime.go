package interpreter

import "sync"

// runtimeState is shared across Evaluator instances (main file, imported modules,
// spawned tasks). It lets us inspect task outcomes at the end of a run.
type runtimeState struct {
	mu    sync.Mutex
	tasks map[*Task]struct{}
}

func newRuntimeState() *runtimeState {
	return &runtimeState{tasks: make(map[*Task]struct{})}
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
