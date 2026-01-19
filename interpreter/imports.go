package interpreter

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"karl/ast"
	"karl/lexer"
	"karl/parser"
)

type moduleState struct {
	mu      sync.Mutex
	loaded  map[string]*Object
	loading map[string]bool
}

func newModuleState() *moduleState {
	return &moduleState{
		loaded:  make(map[string]*Object),
		loading: make(map[string]bool),
	}
}

func (e *Evaluator) evalImportExpression(node *ast.ImportExpression, _ *Environment) (Value, *Signal, error) {
	if node.Path == nil {
		return nil, nil, &RuntimeError{Message: "import expects a string literal path"}
	}
	path, err := e.resolveImportPath(node.Path.Value)
	if err != nil {
		return nil, nil, &RuntimeError{Message: "import path error: " + err.Error()}
	}
	module, err := e.loadModule(path)
	if err != nil {
		return nil, nil, err
	}
	return module, nil, nil
}

func (e *Evaluator) resolveImportPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	root := e.projectRoot
	if root == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = cwd
	}
	return filepath.Join(root, path), nil
}

func (e *Evaluator) loadModule(path string) (*Object, error) {
	resolved := filepath.Clean(path)
	if abs, err := filepath.Abs(resolved); err == nil {
		resolved = abs
	}

	if e.modules == nil {
		e.modules = newModuleState()
	}

	e.modules.mu.Lock()
	if module, ok := e.modules.loaded[resolved]; ok {
		e.modules.mu.Unlock()
		return module, nil
	}
	if e.modules.loading[resolved] {
		e.modules.mu.Unlock()
		return nil, &RuntimeError{Message: "circular import: " + resolved}
	}
	e.modules.loading[resolved] = true
	e.modules.mu.Unlock()

	defer func() {
		e.modules.mu.Lock()
		delete(e.modules.loading, resolved)
		e.modules.mu.Unlock()
	}()

	data, err := os.ReadFile(resolved)
	if err != nil {
		return nil, &RuntimeError{Message: fmt.Sprintf("import read error: %v", err)}
	}
	program, err := parseImportProgram(data, resolved)
	if err != nil {
		return nil, err
	}

	moduleEnv := NewEnclosedEnvironment(NewBaseEnvironment())
	moduleEval := &Evaluator{
		source:      string(data),
		filename:    resolved,
		projectRoot: e.projectRoot,
		modules:     e.modules,
	}
	val, sig, err := moduleEval.Eval(program, moduleEnv)
	if err != nil {
		if _, ok := err.(*RuntimeError); ok {
			return nil, fmt.Errorf("%s", FormatRuntimeError(err, moduleEval.source, moduleEval.filename))
		}
		return nil, err
	}
	if sig != nil {
		return nil, &RuntimeError{Message: "break/continue outside loop"}
	}
	_ = val

	exports := &Object{Pairs: moduleEnv.Snapshot()}
	e.modules.mu.Lock()
	e.modules.loaded[resolved] = exports
	e.modules.mu.Unlock()
	return exports, nil
}

func parseImportProgram(data []byte, filename string) (*ast.Program, error) {
	p := parser.New(lexer.New(string(data)))
	program := p.ParseProgram()
	if errs := p.ErrorsDetailed(); len(errs) > 0 {
		return nil, fmt.Errorf("%s", parser.FormatParseErrors(errs, string(data), filename))
	}
	return program, nil
}
