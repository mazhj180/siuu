package tunnel

import (
	"context"
	"io"
	"net"
	"sync"
	"time"
)

// SystemProxyTunnel system proxy tunnel implementation
type SystemProxyTunnel struct {
	config     *TunnelConfig
	localConn  *monitored
	remoteConn *monitored
	status     *Status
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	isActive   bool
	id         string
}

// NewSystemProxyTunnel create new system proxy tunnel
func NewSystemProxyTunnel(config *TunnelConfig, localConn net.Conn, remoteConn net.Conn, id string) (Tunnel, error) {
	if config == nil {
		config = DefaultTunnelConfig()
	}

	if id == "" {
		id = generateUniqueID()
	}

	ctx, cancel := context.WithCancel(context.Background())

	tunnel := &SystemProxyTunnel{
		id:         id,
		config:     config,
		localConn:  watch(localConn, config.WindowSize),
		remoteConn: watch(remoteConn, config.WindowSize),
		status: &Status{
			ID: id,
		},
		ctx:    ctx,
		cancel: cancel,
	}

	return tunnel, nil
}

func (t *SystemProxyTunnel) ID() string {
	return t.id
}

// Start start the tunnel
func (t *SystemProxyTunnel) Start(ctx context.Context) error {

	if t.isActive {
		return &TunnelError{Type: "already_started", Message: "tunnel is already started"}
	}

	if t.localConn == nil || t.remoteConn == nil {
		return &TunnelError{Type: "connection_not_established", Message: "connection not established"}
	}

	t.isActive = true
	t.status.StartTime = time.Now()

	// start bidirectional data forwarding
	return t.forwardData()
}

// forwardData forward data
func (t *SystemProxyTunnel) forwardData() error {
	errCh := make(chan error, 2)

	// goroutine 1: local -> remote (upstream)
	go func() {
		defer func() {
			if c, ok := t.remoteConn.ReadWriteCloser.(interface{ CloseWrite() error }); ok {
				_ = c.CloseWrite()
			}
		}()

		err := t.copy(t.remoteConn, t.localConn)
		errCh <- err
	}()

	// goroutine 2: remote -> local (downstream)
	go func() {
		defer func() {
			if c, ok := t.localConn.ReadWriteCloser.(interface{ CloseWrite() error }); ok {
				_ = c.CloseWrite()
			}
		}()

		err := t.copy(t.localConn, t.remoteConn)
		errCh <- err
	}()

	// wait for two goroutines to complete
	err1 := <-errCh
	err2 := <-errCh

	// update status
	t.mu.Lock()
	t.isActive = false
	t.status.EndTime = time.Now()
	t.status.TotalDuration = t.status.EndTime.Sub(t.status.StartTime)
	t.status.UpBytes = t.localConn.totalBytes
	t.status.DownBytes = t.remoteConn.totalBytes
	_, t.status.UpSpeed = t.localConn.calculateSpeed()
	_, t.status.DownSpeed = t.remoteConn.calculateSpeed()

	t.mu.Unlock()

	// close connection
	t.Close()

	// record error (if any)
	if err1 != nil && err1 != io.EOF {
		return err1
	}
	if err2 != nil && err2 != io.EOF {
		return err2
	}
	return nil
}

// copyWithMonitor copy data with monitor
func (t *SystemProxyTunnel) copy(dst *monitored, src *monitored) error {
	buffer := make([]byte, t.config.BufferSize)
	var err error
	if _, err = io.CopyBuffer(dst, src, buffer); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// Close close the tunnel
func (t *SystemProxyTunnel) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.cancel()
	t.isActive = false
	t.status.EndTime = time.Now()
	t.status.TotalDuration = t.status.EndTime.Sub(t.status.StartTime)

	if t.localConn != nil {
		t.localConn.Close()
	}
	if t.remoteConn != nil {
		t.remoteConn.Close()
	}

	return nil
}

// Ping test connection delay
func (t *SystemProxyTunnel) Ping() (time.Duration, error) {
	if !t.isActive {
		return 0, &TunnelError{Type: "not_active", Message: "tunnel is not active"}
	}

	start := time.Now()

	// send ping packet (here simplified, actually can send specific ping packet)
	if t.remoteConn != nil {
		if _, err := t.remoteConn.Write([]byte("ping")); err != nil {
			return 0, err
		}
	}

	return time.Since(start), nil
}

// GetStatus get the tunnel status
func (t *SystemProxyTunnel) GetStatus() *Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

func (t *SystemProxyTunnel) CalculateUpSpeed() (currentSpeed, averageSpeed float64) {
	return t.localConn.calculateSpeed()
}

func (t *SystemProxyTunnel) CalculateDownSpeed() (currentSpeed, averageSpeed float64) {
	return t.remoteConn.calculateSpeed()
}
