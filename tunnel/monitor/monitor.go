package monitor

import (
	"net"
	"time"
)

type Interface interface {
	net.Conn
	UpTraffic() (int64, float64)
	DownTraffic() (int64, float64)
}

func Watch(conn net.Conn) Interface {
	return &monitor{
		Conn:      conn,
		upTimer:   timer{},
		downTimer: timer{},
	}
}

type monitor struct {
	net.Conn
	up        int64
	down      int64
	upTimer   timer
	downTimer timer
}

func (m *monitor) Write(b []byte) (int, error) {
	m.downTimer.start()
	n, err := m.Conn.Write(b)
	if n > 0 {
		m.down += int64(n)
	}
	m.downTimer.stop()
	return n, err
}

func (m *monitor) Read(b []byte) (int, error) {
	m.upTimer.start()
	n, err := m.Conn.Read(b)
	if n > 0 {
		m.up += int64(n)
	}
	m.upTimer.stop()
	return n, err
}

func (m *monitor) UpTraffic() (int64, float64) {
	up := float64(m.up)
	upCost := m.upTimer.cost()
	speed := up / upCost
	return m.up, speed
}

func (m *monitor) DownTraffic() (int64, float64) {
	down := float64(m.down)
	downCost := m.downTimer.cost()
	speed := down / downCost
	return m.down, speed
}

type timer struct {
	startTime time.Time
	duration  time.Duration
}

func (t *timer) start() {
	if t.startTime.IsZero() {
		t.startTime = time.Now()
	}
}

func (t *timer) stop() {
	if !t.startTime.IsZero() {
		t.duration += time.Since(t.startTime)
		t.startTime = time.Time{}
	}
}

func (t *timer) cost() float64 {
	return t.duration.Seconds()
}
