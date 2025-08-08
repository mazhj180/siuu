package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TimeSplitter time interval splitter
type TimeSplitter struct {
	interval  time.Duration
	lastSplit time.Time
}

// NewTimeSplitter create time interval splitter
func NewTimeSplitter(interval time.Duration) Splitter {
	return &TimeSplitter{
		interval:  interval,
		lastSplit: time.Now(),
	}
}

func (t *TimeSplitter) ShouldSplit(file *os.File, message string) bool {
	now := time.Now()
	if now.Sub(t.lastSplit) >= t.interval {
		t.lastSplit = now
		return true
	}
	return false
}

func (t *TimeSplitter) GetNextFileName(baseDir, baseName string) string {
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(baseDir, fmt.Sprintf("%s-%s.log", baseName, timestamp))
}

func (t *TimeSplitter) Cleanup(baseDir, baseName string) error {
	// time interval splitter does not need to be cleaned up, keep all files
	return nil
}

// CompositeSplitter composite splitter, support multiple split strategies
type CompositeSplitter struct {
	splitters []Splitter
}

// NewCompositeSplitter create composite splitter
func NewCompositeSplitter(splitters ...Splitter) *CompositeSplitter {
	return &CompositeSplitter{
		splitters: splitters,
	}
}

func (c *CompositeSplitter) ShouldSplit(file *os.File, message string) bool {
	for _, splitter := range c.splitters {
		if splitter.ShouldSplit(file, message) {
			return true
		}
	}
	return false
}

func (c *CompositeSplitter) GetNextFileName(baseDir, baseName string) string {
	// use the naming strategy of the first splitter
	if len(c.splitters) > 0 {
		return c.splitters[0].GetNextFileName(baseDir, baseName)
	}
	// use timestamp by default
	timestamp := time.Now().Format("20060102-150405")
	return filepath.Join(baseDir, fmt.Sprintf("%s-%s.log", baseName, timestamp))
}

func (c *CompositeSplitter) Cleanup(baseDir, baseName string) error {
	// execute cleanup operation for all splitters
	for _, splitter := range c.splitters {
		if err := splitter.Cleanup(baseDir, baseName); err != nil {
			return err
		}
	}
	return nil
}
