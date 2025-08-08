package splitter

import (
	"fmt"
	"os"
	"path/filepath"
)

// SizeOverwriteSplitter size splitter that splits by size and directly overwrites the file
type SizeOverwriteSplitter struct {
	maxSize int64 // maximum file size (bytes)
}

// NewSizeOverwriteSplitter create size splitter that splits by size and directly overwrites the file
func NewSizeOverwriteSplitter(maxSize int64) Splitter {
	return &SizeOverwriteSplitter{
		maxSize: maxSize,
	}
}

// ShouldSplit check if the file needs to be split
func (s *SizeOverwriteSplitter) ShouldSplit(file *os.File, message string) bool {
	if file == nil {
		return false
	}

	// get current file size
	stat, err := file.Stat()
	if err != nil {
		return false
	}

	// calculate new message size (including newline)
	newMessageSize := int64(len(message) + 1)

	// if current size plus new message size exceeds maximum size, then need to split
	return stat.Size()+newMessageSize > s.maxSize
}

// GetNextFileName get next file name (overwrite mode, always use the same file name)
func (s *SizeOverwriteSplitter) GetNextFileName(baseDir, baseName string) string {
	return filepath.Join(baseDir, fmt.Sprintf("%s.log", baseName))
}

// Cleanup clean up old files (overwrite mode does not need to be cleaned up)
func (s *SizeOverwriteSplitter) Cleanup(baseDir, baseName string) error {
	// overwrite mode does not need to be cleaned up, return directly
	return nil
}

// SetMaxSize set maximum file size
func (s *SizeOverwriteSplitter) SetMaxSize(maxSize int64) {
	s.maxSize = maxSize
}

// GetMaxSize get maximum file size
func (s *SizeOverwriteSplitter) GetMaxSize() int64 {
	return s.maxSize
}
