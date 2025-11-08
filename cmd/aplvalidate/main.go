package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"wotlk-destro-sim/internal/apl"
)

func main() {
	var rotationPath string
	flag.StringVar(&rotationPath, "rotation", "configs/rotations/destruction-default.yaml", "Path to rotation YAML")
	flag.Parse()

	rotationPath = filepath.Clean(rotationPath)
	baseDir := filepath.Dir(rotationPath)
	rel := filepath.Base(rotationPath)

	file, err := apl.LoadRotation(baseDir, rel)
	if err != nil {
		log.Fatalf("failed to load rotation: %v", err)
	}

	if _, err := apl.Compile(file); err != nil {
		log.Fatalf("rotation invalid: %v", err)
	}

	fmt.Printf("Rotation '%s' validated successfully (source: %s)\n", file.Name, rotationPath)
}
