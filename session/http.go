package session

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"siuu/server/store"
	"siuu/tunnel/proto"
	"siuu/tunnel/proxy"
	"strconv"
	"strings"
)

type httpSession struct {
	prx                proxy.Proxy
	conn               net.Conn
	isTLS              bool
	addr               *Addr
	id                 string
	buf                []byte
	up, down           int64
	upSpeed, downSpeed float64
}

func OpenHttpSession(conn net.Conn) Session {
	sid := "h-" + genSid()
	return &httpSession{
		conn: conn,
		id:   sid,
		prx:  store.GetDirect(),
	}
}

func (s *httpSession) Handshakes() error {
	req, err := http.ReadRequest(bufio.NewReader(s.conn))
	if err != nil {
		return err
	}
	var domain string
	var port uint64 = 80
	if strings.Contains(req.Host, ":") {
		d, p, _ := net.SplitHostPort(req.Host)
		domain = d
		port, _ = strconv.ParseUint(p, 10, 16)
	} else {
		domain = req.Host
	}

	s.addr = &Addr{
		Domain: domain,
		Port:   uint16(port),
	}
	if req.Method == http.MethodConnect {
		s.isTLS = true
		s.addr.Port = 443
		if _, err = s.conn.Write([]byte(fmt.Sprintf("%s 200 Connection Established\r\n\r\n", req.Proto))); err != nil {
			return err
		}
	} else {
		var buf bytes.Buffer
		path := req.URL.Path
		if req.URL.RawQuery != "" {
			path += "?" + req.URL.RawQuery
		}
		_, _ = fmt.Fprintf(&buf, "%s %s %s\r\n", req.Method, path, req.Proto)

		hasHost := false
		for k, vals := range req.Header {
			if strings.EqualFold(k, "Host") {
				hasHost = true
			}
			for _, val := range vals {
				_, _ = fmt.Fprintf(&buf, "%s: %s\r\n", k, val)
			}
		}

		if !hasHost && req.Host != "" {
			_, _ = fmt.Fprintf(&buf, "Host: %s\r\n", req.Host)
		}

		_, _ = fmt.Fprintf(&buf, "\r\n")
		s.buf = buf.Bytes()
	}
	return nil
}

func (s *httpSession) String() string {
	return fmt.Sprintf("%s://%s:%d", s.addr.Domain, s.addr.Port, s.addr.Port)
}

func (s *httpSession) GetHost() string {
	if s.addr.Domain != "" {
		return s.addr.Domain
	}
	return s.addr.IP.String()
}

func (s *httpSession) GetPort() uint16 {
	return s.addr.Port
}

func (s *httpSession) GetProtocol() proto.Protocol {
	return proto.HTTP
}

func (s *httpSession) GetProxy() proxy.Proxy {
	return s.prx
}

func (s *httpSession) GetConn() net.Conn {
	return s.conn
}

func (s *httpSession) ID() string {
	return s.id
}

func (s *httpSession) SetProxy(p proxy.Proxy) {
	s.prx = p
}

func (s *httpSession) IsTLS() bool {
	return s.isTLS
}

func (s *httpSession) GetOtherData() []byte {
	return s.buf
}

func (s *httpSession) RecordUp(up int64, speed float64) {
	s.up = up
	s.upSpeed = speed
}

func (s *httpSession) RecordDown(down int64, speed float64) {
	s.down = down
	s.downSpeed = speed
}
