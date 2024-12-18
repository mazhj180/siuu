package session

import (
	"bufio"
	"evil-gopher/proxy"
	"evil-gopher/tunnel"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type httpSession struct {
	prx   proxy.Proxy
	conn  net.Conn
	isTLS bool
	addr  *Addr
	id    string
}

func OpenHttpSession(conn net.Conn) Session {
	sid := "h-" + genSid()
	return &httpSession{
		conn: conn,
		id:   sid,
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

func (s *httpSession) GetProtocol() tunnel.Protocol {
	return tunnel.HTTP
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
