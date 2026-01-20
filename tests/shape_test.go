package tests

import (
	"path/filepath"
	"testing"

	"karl/ast"
)

func TestExampleShapes(t *testing.T) {
	t.Run("03_loops_and_functions", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "loops_and_functions.k"))
		requireCountAtLeast(t, program, "ForExpression", 3, func(n ast.Node) bool {
			_, ok := n.(*ast.ForExpression)
			return ok
		})
		requireCountAtLeast(t, program, "BreakExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.BreakExpression)
			return ok
		})
		requireCountAtLeast(t, program, "ContinueExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.ContinueExpression)
			return ok
		})
	})

	t.Run("05_concurrency", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "concurrency.k"))
		requireCountAtLeast(t, program, "SpawnExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.SpawnExpression)
			return ok
		})
		requireCountAtLeast(t, program, "RaceExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.RaceExpression)
			return ok
		})
		requireCountAtLeast(t, program, "ForExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.ForExpression)
			return ok
		})
	})

	t.Run("08_full_program", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "full_program.k"))
		requireCountAtLeast(t, program, "QueryExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.QueryExpression)
			return ok
		})
		requireCountAtLeast(t, program, "RangeExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.RangeExpression)
			return ok
		})
	})
}

func TestExerciseShapes(t *testing.T) {
	t.Run("01_pulse_pipeline", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "pulse_pipeline.k"))
		requireCountAtLeast(t, program, "SpawnExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.SpawnExpression)
			return ok
		})
	})

	t.Run("05_game_tick", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "game_tick.k"))
		requireCountAtLeast(t, program, "ForExpression", 2, func(n ast.Node) bool {
			_, ok := n.(*ast.ForExpression)
			return ok
		})
	})

	t.Run("08_retry_policy", func(t *testing.T) {
		program := parseFile(t, filepath.Join("..", "examples", "programs", "retry_policy.k"))
		requireCountAtLeast(t, program, "BreakExpression", 1, func(n ast.Node) bool {
			_, ok := n.(*ast.BreakExpression)
			return ok
		})
	})
}
