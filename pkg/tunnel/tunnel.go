package tunnel

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// Status tunnel status information
type Status struct {
	ID         string
	ClientName string

	UpBytes       int64
	DownBytes     int64
	UpSpeed       float64
	DownSpeed     float64
	TotalDuration time.Duration
	StartTime     time.Time
	EndTime       time.Time
}

// Tunnel tunnel interface
type Tunnel interface {
	ID() string
	Start(ctx context.Context) error // Start open the tunnel
	Close() error                    // Close close the tunnel
	Ping() (time.Duration, error)    // Ping test the connection delay

	GetStatus() *Status // GetStatus get the tunnel status

	// CalculateUpSpeed calculate the up speed and average speed
	CalculateUpSpeed() (currentSpeed, averageSpeed float64)

	// CalculateDownSpeed calculate the down speed and average speed
	CalculateDownSpeed() (currentSpeed, averageSpeed float64)
}

// TunnelConfig 隧道配置
type TunnelConfig struct {
	BufferSize     int
	WindowSize     int
	UpdateInterval time.Duration
	Timeout        time.Duration
}

// DefaultTunnelConfig 默认隧道配置
func DefaultTunnelConfig() *TunnelConfig {
	return &TunnelConfig{
		BufferSize:     2 * 1024, // 2KB
		WindowSize:     10,
		UpdateInterval: time.Second,
		Timeout:        30 * time.Second,
	}
}

// TunnelError 隧道错误
type TunnelError struct {
	Type    string
	Message string
}

func (e *TunnelError) Error() string {
	return e.Message
}

func generateUniqueID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf("%x%x", timestamp, randomBytes)
}
