package splitter

import (
	"os"
)

// Splitter splitter interface
type Splitter interface {
	// ShouldSplit check if the file needs to be split
	ShouldSplit(file *os.File, message string) bool

	// GetNextFileName get next file name
	GetNextFileName(baseDir, baseName string) string

	// Cleanup clean up old files (optional)
	Cleanup(baseDir, baseName string) error
}
