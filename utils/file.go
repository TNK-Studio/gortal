package utils

import (
	"os"
	"strings"
)

// FilePath replace ~ -> $HOME
func FilePath(path string) string {
	path = strings.Replace(path, "~", os.Getenv("HOME"), 1)
	return path
}
