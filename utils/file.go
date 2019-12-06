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

// FileExited check file exited
func FileExited(path string) bool {
	info, err := os.Stat(FilePath(path))
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// IsDirector IsDir
func IsDirector(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
