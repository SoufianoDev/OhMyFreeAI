package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func moveRestoreFileNoReplace(source, destination string) error {
	if err := validateRestoreMovePaths(source, destination); err != nil {
		return err
	}
	if err := moveRestoreFileNoReplacePlatform(source, destination); err != nil {
		return fmt.Errorf("move SQLite restore artifact without replacement: %w", err)
	}
	return nil
}

func validateRestoreMovePaths(source, destination string) error {
	if strings.TrimSpace(source) == "" || strings.TrimSpace(destination) == "" {
		return fmt.Errorf("SQLite restore move paths must not be empty")
	}
	source, err := filepath.Abs(filepath.Clean(source))
	if err != nil {
		return fmt.Errorf("resolve SQLite restore source: %w", err)
	}
	destination, err = filepath.Abs(filepath.Clean(destination))
	if err != nil {
		return fmt.Errorf("resolve SQLite restore destination: %w", err)
	}
	if equalRestoreMovePath(source, destination) {
		return fmt.Errorf("SQLite restore source and destination are identical")
	}
	sourceDirectory := filepath.Clean(filepath.Dir(source))
	destinationDirectory := filepath.Clean(filepath.Dir(destination))
	if !equalRestoreMovePath(sourceDirectory, destinationDirectory) {
		return fmt.Errorf("SQLite restore move must stay within one directory")
	}
	info, err := os.Lstat(source)
	if err != nil {
		return fmt.Errorf("inspect SQLite restore source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("SQLite restore source is not a regular file")
	}
	return nil
}

func equalRestoreMovePath(left, right string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
