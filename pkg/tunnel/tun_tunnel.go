package tunnel

import (
	"context"
	"sync"
	"time"
)

// tunTunnel TUN mode tunnel implementation (placeholder)
type tunTunnel struct {
	config      *TunnelConfig
	upMonitor   *monitored
	downMonitor *monitored
	status      *Status
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	isActive    bool
	id          string
}

// NewTunTunnel create new TUN tunnel
func NewTunTunnel(config *TunnelConfig, id string) (Tunnel, error) {
	if config == nil {
		config = DefaultTunnelConfig()
	}

	if id == "" {
		id = generateUniqueID()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &tunTunnel{
		id:          id,
		config:      config,
		upMonitor:   watch(nil, config.WindowSize),
		downMonitor: watch(nil, config.WindowSize),
		status: &Status{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return tunnel, nil
}

func (t *tunTunnel) ID() string {
	return t.id
}

// Start start the tunnel
func (t *tunTunnel) Start(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isActive {
		return &TunnelError{Type: "already_started", Message: "tunnel is already started"}
	}

	// TODO: implement TUN mode specific logic
	// 1. create TUN device
	// 2. configure network interface
	// 3. start data packet processing
	// 4. implement IP packet routing and forwarding

	t.isActive = true
	t.status.StartTime = time.Now()

	return &TunnelError{Type: "not_implemented", Message: "TUN tunnel is not implemented yet"}
}

// Close close the tunnel
func (t *tunTunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()
	t.isActive = false
	t.status.EndTime = time.Now()
	t.status.TotalDuration = t.status.EndTime.Sub(t.status.StartTime)

	// TODO: clean up TUN device resources

	return nil
}

// Ping test connection delay
func (t *tunTunnel) Ping() (time.Duration, error) {
	if !t.isActive {
		return 0, &TunnelError{Type: "not_active", Message: "tunnel is not active"}
	}

	// TODO: implement TUN mode ping logic
	return 0, &TunnelError{Type: "not_implemented", Message: "TUN tunnel ping is not implemented yet"}
}

// GetStatus get the tunnel status
func (t *tunTunnel) GetStatus() *Status {
	t.mu.RLock()
	defer t.mu.RUnlock()

	status := *t.status
	return &status
}

func (t *tunTunnel) CalculateUpSpeed() (currentSpeed, averageSpeed float64) {
	return t.upMonitor.calculateSpeed()
}

func (t *tunTunnel) CalculateDownSpeed() (currentSpeed, averageSpeed float64) {
	return t.downMonitor.calculateSpeed()
}
