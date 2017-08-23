package main

import (
	"fmt"
	"path/filepath"
)

type badGlobPatternError struct {
	pattern string
}

func findFiles(globPattern string) ([]string, error) {
	return filepath.Glob(globPattern)
}

func (e *badGlobPatternError) Error() string {
	message := fmt.Sprintf("The pattern \"%s\" is not a valid glob pattern", e.pattern)
	return message
}
