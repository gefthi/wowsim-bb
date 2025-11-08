package apl

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// TODO: add CLI validator that uses this loader to lint rotations before runtime.
// (See roadmap step 1.)

// LoadRotation loads a rotation file from configs/rotations, resolving imports.
func LoadRotation(baseDir, relPath string) (*File, error) {
	seen := map[string]bool{}
	return loadRecursive(baseDir, relPath, seen)
}

func loadRecursive(baseDir, relPath string, seen map[string]bool) (*File, error) {
	normalized := filepath.Clean(relPath)
	if seen[normalized] {
		return nil, fmt.Errorf("rotation import cycle detected at %s", normalized)
	}
	seen[normalized] = true

	fullPath := filepath.Join(baseDir, normalized)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}

	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse %s: %w", relPath, err)
	}

	// Resolve imports depth-first.
	var compiledRotation []ActionDefinition
	for _, imp := range file.Imports {
		child, err := loadRecursive(baseDir, imp, seen)
		if err != nil {
			return nil, err
		}
		compiledRotation = append(compiledRotation, child.Rotation...)
	}
	compiledRotation = append(compiledRotation, file.Rotation...)
	file.Rotation = compiledRotation

	seen[normalized] = false
	return &file, nil
}
