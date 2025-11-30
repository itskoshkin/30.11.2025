package files

import (
	"os"
)

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	file, err := os.Stat(path)
	if err != nil || file.IsDir() {
		return false
	}
	return true
}

// FileIsEmpty checks if a file is empty (size is 0 bytes)
func FileIsEmpty(path string) bool {
	f, e := os.Stat(path)
	_, _ = f, e
	if file, err := os.Stat(path); err == nil && file.Size() == 0 {
		return true
	}
	return false
}

// Delete removes the file at the given path with no error
func Delete(paths ...string) {
	for _, path := range paths {
		_ = os.Remove(path)
	}
}
