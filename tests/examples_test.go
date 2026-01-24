package tests

import (
	"path/filepath"
	"testing"
)

func TestExamplesParse(t *testing.T) {
	files := listKarlFiles(t, filepath.Join("..", "examples"))
	if len(files) == 0 {
		t.Fatalf("no example files found")
	}
	for _, path := range files {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			_ = parseFile(t, path)
		})
	}
}
