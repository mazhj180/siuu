package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DateSplitter struct {
	dateFormat  string
	currentDate string
}

// NewDateSplitter create date splitter
// @param dateFormat date format
// @return Splitter date splitter
func NewDateSplitter(dateFormat string) Splitter {
	return &DateSplitter{
		dateFormat:  dateFormat,
		currentDate: time.Now().Format(dateFormat),
	}
}

func (d *DateSplitter) ShouldSplit(file *os.File, message string) bool {
	currentDate := time.Now().Format(d.dateFormat)
	if d.currentDate != currentDate {
		d.currentDate = currentDate
		return true
	}
	return false
}

func (d *DateSplitter) GetNextFileName(baseDir, baseName string) string {
	return filepath.Join(baseDir, fmt.Sprintf("%s-%s.log", baseName, d.currentDate))
}

func (d *DateSplitter) Cleanup(baseDir, baseName string) error {
	// date splitter does not need to be cleaned up, keep historical files
	return nil
}

// SizeSplitter size splitter
type SizeSplitter struct {
	maxSize   int64 // maximum file size (bytes)
	maxFiles  int   // maximum file number
	overwrite bool  // overwrite mode
}

// NewSizeSplitter create size splitter
func NewSizeSplitter(maxSize int64, maxFiles int, overwrite bool) *SizeSplitter {
	return &SizeSplitter{
		maxSize:   maxSize,
		maxFiles:  maxFiles,
		overwrite: overwrite,
	}
}

func (s *SizeSplitter) ShouldSplit(file *os.File, message string) bool {
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

func (s *SizeSplitter) GetNextFileName(baseDir, baseName string) string {
	if s.overwrite {
		// overwrite mode: always use the same file name
		return filepath.Join(baseDir, fmt.Sprintf("%s.log", baseName))
	}

	// non-overwrite mode: use timestamp to generate unique file name
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(baseDir, fmt.Sprintf("%s-%s.log", baseName, timestamp))
}

func (s *SizeSplitter) Cleanup(baseDir, baseName string) error {
	if s.overwrite {
		// overwrite mode does not need to be cleaned up
		return nil
	}

	// get all matching files
	pattern := filepath.Join(baseDir, fmt.Sprintf("%s-*.log", baseName))
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// if file number exceeds maximum number, delete the oldest file
	if len(files) > s.maxFiles {
		// sort by modification time
		type fileInfo struct {
			path    string
			modTime time.Time
		}

		var fileInfos []fileInfo
		for _, file := range files {
			stat, err := os.Stat(file)
			if err != nil {
				continue
			}
			fileInfos = append(fileInfos, fileInfo{
				path:    file,
				modTime: stat.ModTime(),
			})
		}

		// sort by modification time (oldest first)
		for i := 0; i < len(fileInfos)-1; i++ {
			for j := i + 1; j < len(fileInfos); j++ {
				if fileInfos[i].modTime.After(fileInfos[j].modTime) {
					fileInfos[i], fileInfos[j] = fileInfos[j], fileInfos[i]
				}
			}
		}

		// delete the oldest file
		toDelete := len(fileInfos) - s.maxFiles
		for i := 0; i < toDelete; i++ {
			os.Remove(fileInfos[i].path)
		}
	}

	return nil
}
