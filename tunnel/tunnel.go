package tunnel

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"siuu/logger"
	"siuu/tunnel/monitor"
	"siuu/tunnel/proto"
	"siuu/tunnel/proxy"
	"siuu/tunnel/tester"
	"sync"
	"time"
)

var (
	T              Tunnel
	PingTimeoutErr = errors.New("ping timeout")
)

type Tunnel interface {
	In(proto.Interface) (Traffic, error)
	Interrupt() []string
	Ping(proxy.Proxy) (Traffic, error)
}

func init() {
	T = &tunnel{
		livelyConn: make(map[string]net.Conn),
	}
}

type tunnel struct {
	rwx        sync.RWMutex
	livelyConn map[string]net.Conn
}

// In starts a tunnel from the given connection to the given host and port.
//
// It first registers the connection to the tunnel, then dials to the proxy.
// If the dialing fails, it returns immediately.
//
// After that, it copies data from the connection to the proxy and from the
// proxy to the connection. If either of the copies fails, it logs the error
// and closes the connection.
//
// Finally, it records the traffic and logs it.
func (t *tunnel) In(p proto.Interface) (Traffic, error) {
	sid := p.ID()
	host, port, conn, prx := p.GetHost(), p.GetPort(), p.GetConn(), p.GetProxy()

	t.register(sid, conn)

	// dial to proxy
	timer := monitor.Timer{}
	timer.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agency, err := prx.Connect(ctx, host, port)
	if err != nil {
		return Traffic{}, err
	}
	delay := timer.Cost()

	// forward
	monitored := monitor.Watch(conn)

	if prx.GetType() == proxy.DIRECT {
		h, ok := p.(proto.HttpInterface)
		if ok {
			reader := h.GetHttpReader()
			if reader != nil {
				n, _ := io.Copy(agency, reader)
				monitored.Extra(n, 0)
			}
		}
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer agency.CloseWriter()
		defer t.remove(sid)
		if _, err = io.Copy(agency, monitored); err != nil {
			logger.SWarn("<%s> %s", sid, err)
		}

	}()

	if _, err = io.Copy(monitored, agency); err != nil {
		logger.SWarn("<%s> %s", sid, err)
	}
	_ = monitored.CloseWriter()

	<-done
	_ = monitored.Close()
	_ = agency.Close()

	// recorded
	up, upSpeed := monitored.UpTraffic()
	down, downSpeed := monitored.DownTraffic()
	upSpend, downSpend := monitored.SpendTime()

	if rt, ok := p.(proto.TrafficRecorder); ok {
		rt.RecordUp(up, upSpeed)
		rt.RecordDown(down, downSpeed)
	}

	tr := Traffic{
		Up:            up,
		Down:          down,
		UpSpendTime:   upSpend,
		DownSpendTime: downSpend,
		UpSpeed:       upSpeed,
		DownSpeed:     downSpeed,
		Delay:         delay,
	}

	return tr, err

}

// Interrupt closes all established connections and returns the session ids of
// the interrupted connections.
//
// This function is useful for stopping the tunnel and closing all connections gracefully.
func (t *tunnel) Interrupt() []string {
	t.rwx.Lock()
	defer t.rwx.Unlock()

	l := len(t.livelyConn)
	sids := make([]string, l)
	for sid, conn := range t.livelyConn {
		_ = conn.Close()
		sids = append(sids, sid)
	}

	t.livelyConn = make(map[string]net.Conn)
	return sids
}

// Ping pings the given proxy and returns the traffic.
//
// It first creates a test connection, then dials to the proxy.
// If the dialing fails, it returns immediately.
//
// After that, it copies data from the test connection to the proxy and from the
// proxy to the test connection. If either of the copies fails, it logs the error
// and closes the connection.
//
// Finally, it records the traffic and logs it.
func (t *tunnel) Ping(prx proxy.Proxy) (Traffic, error) {
	host := "github.com"
	if prx.GetType() == proxy.DIRECT {
		host = "baidu.com"
	}
	req, err := http.NewRequest("GET", "https://"+host, nil)
	if err != nil {
		return Traffic{}, err
	}

	r := proxy.NewHttpReader(req)
	w := &bytes.Buffer{}

	testConn := &tester.TestConn{
		Reader: r,
		Writer: w,
	}
	monitored := monitor.Watch(testConn)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// init timer
	timer := monitor.Timer{}
	timer.Start()

	var agency *proxy.Pd
	done := make(chan struct{})
	go func() {
		defer close(done)
		agency, err = prx.Connect(ctx, host, 443)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return Traffic{}, PingTimeoutErr
	}

	if err != nil {
		return Traffic{}, err
	}

	delay := timer.Cost()

	done = make(chan struct{})
	go func() {
		defer close(done)
		defer agency.CloseWriter()
		_, _ = io.Copy(agency, monitored)
	}()

	go func() {
		_, _ = io.Copy(monitored, agency)
	}()

	var tr Traffic
	select {
	case <-done:
		tr.Up, tr.UpSpeed = monitored.UpTraffic()
		tr.Down, tr.DownSpeed = monitored.DownTraffic()
		tr.UpSpendTime, tr.DownSpendTime = monitored.SpendTime()
		tr.Delay = delay
	case <-ctx.Done():
		err = PingTimeoutErr
	}
	_ = agency.Close()

	return tr, err
}

// register records a connection to the tunnel, by the given sid.
//
// This function will be called when a connection is established.
func (t *tunnel) register(sid string, conn net.Conn) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	t.livelyConn[sid] = conn
}

func (t *tunnel) remove(sid string) {
	t.rwx.Lock()
	defer t.rwx.Unlock()
	if _, ok := t.livelyConn[sid]; ok {
		delete(t.livelyConn, sid)
	}
}

type Traffic struct {
	Up, Down                   int64
	UpSpendTime, DownSpendTime float64
	UpSpeed, DownSpeed         float64

	Delay float64
}
