package monitor

import (
	"errors"
	"net"
	"time"
)

type Interface interface {
	net.Conn
	CloseWriter() error
	CloseReader() error
	UpTraffic() (int64, float64)
	DownTraffic() (int64, float64)
	SpendTime() (float64, float64)
	Extra(int64, int64)
}

func Watch(conn net.Conn, initValues ...int64) Interface {

	var initV int64
	for i, _ := range initValues {
		initV += initValues[i]
	}
	return &monitor{
		Conn:      conn,
		up:        initV,
		upTimer:   Timer{},
		downTimer: Timer{},
	}
}

type monitor struct {
	net.Conn
	up        int64
	down      int64
	upTimer   Timer
	downTimer Timer
}

func (m *monitor) CloseWriter() error {
	if conn, ok := m.Conn.(*net.TCPConn); ok {
		return conn.CloseWrite()
	}
	return errors.New("it is not tcp conn")
}

func (m *monitor) CloseReader() error {
	if conn, ok := m.Conn.(*net.TCPConn); ok {
		return conn.CloseRead()
	}
	return errors.New("it is not tcp conn")
}

func (m *monitor) Write(b []byte) (int, error) {
	m.downTimer.Start()
	n, err := m.Conn.Write(b)
	if n > 0 {
		m.down += int64(n)
	}
	m.downTimer.Stop()
	return n, err
}

func (m *monitor) Read(b []byte) (int, error) {
	m.upTimer.Start()
	n, err := m.Conn.Read(b)
	if n > 0 {
		m.up += int64(n)
	}
	m.upTimer.Stop()
	return n, err
}

func (m *monitor) UpTraffic() (int64, float64) {
	up := float64(m.up)
	upCost := m.upTimer.Cost()
	speed := up / upCost
	return m.up, speed
}

func (m *monitor) DownTraffic() (int64, float64) {
	down := float64(m.down)
	downCost := m.downTimer.Cost()
	speed := down / downCost
	return m.down, speed
}

func (m *monitor) SpendTime() (up, down float64) {
	up = m.upTimer.Cost()
	down = m.downTimer.Cost()
	return
}

func (m *monitor) Extra(up, down int64) {
	m.up += up
	m.down += down
}

type Timer struct {
	startTime time.Time
	duration  time.Duration
}

func (t *Timer) Start() {
	if t.startTime.IsZero() {
		t.startTime = time.Now()
	}
}

func (t *Timer) Stop() {
	if !t.startTime.IsZero() {
		t.duration += time.Since(t.startTime)
		t.startTime = time.Time{}
	}
}

func (t *Timer) Cost() float64 {
	if !t.startTime.IsZero() {
		return time.Since(t.startTime).Seconds()
	}
	return t.duration.Seconds()
}
