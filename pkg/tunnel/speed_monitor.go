package tunnel

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type speedEntry struct {
	timestamp time.Time
	bytes     int64
}

type monitored struct {
	io.ReadWriteCloser
	mu         sync.RWMutex
	window     []speedEntry
	maxWindow  int
	totalBytes int64

	startTime time.Time
}

func watch(reader io.ReadWriteCloser, maxWindow int) *monitored {
	return &monitored{
		ReadWriteCloser: reader,
		maxWindow:       maxWindow,
	}

}

func (m *monitored) Read(p []byte) (n int, err error) {

	if m.startTime.IsZero() {
		m.startTime = time.Now()
	}

	n, err = m.ReadWriteCloser.Read(p)
	now := time.Now()

	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.window) >= m.maxWindow {
		m.window = m.window[1:]
	}

	m.window = append(m.window, speedEntry{timestamp: now, bytes: int64(n)})

	m.totalBytes += int64(n)

	return n, err
}

func (m *monitored) calculateSpeed() (currentSpeed, averageSpeed float64) {

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.window) == 0 {
		return
	}

	// calculate current speed (based on the latest data point)
	if len(m.window) >= 2 {
		last := m.window[len(m.window)-1]
		prev := m.window[len(m.window)-2]
		timeDiff := last.timestamp.Sub(prev.timestamp).Seconds()
		if timeDiff > 0 {
			currentSpeed = float64(last.bytes) / timeDiff
		}
	}

	// calculate average speed (based on all data in the window)
	totalBytes := int64(0)
	for _, data := range m.window {
		totalBytes += data.bytes
	}

	if len(m.window) > 0 {
		first := m.window[0]
		last := m.window[len(m.window)-1]
		timeSpan := last.timestamp.Sub(first.timestamp).Seconds()
		if timeSpan > 0 {
			averageSpeed = float64(totalBytes) / timeSpan
		}
	}

	return currentSpeed, averageSpeed
}

// FormatSpeed format speed display
func FormatSpeed(speed float64) string {
	const (
		KB = 1024
		MB = 1024 * 1024
		GB = 1024 * 1024 * 1024
	)

	switch {
	case speed >= GB:
		return fmt.Sprintf("%.2f GB/s", speed/GB)
	case speed >= MB:
		return fmt.Sprintf("%.2f MB/s", speed/MB)
	case speed >= KB:
		return fmt.Sprintf("%.2f KB/s", speed/KB)
	default:
		return fmt.Sprintf("%.0f B/s", speed)
	}
}
