package session

import (
	"bufio"
	"evil-gopher/proxy"
	"evil-gopher/routing"
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
	host  *routing.TargetHost
	id    string
}

func OpenHttpSession(conn net.Conn) Session {
	sid := "h-" + genSid()
	return &httpSession{
		conn: conn,
		id:   sid,
	}
}

func (s *httpSession) Handshakes() (*routing.TargetHost, error) {
	req, err := http.ReadRequest(bufio.NewReader(s.conn))
	if err != nil {
		return nil, err
	}
	var domain string
	port := 80
	if strings.Contains(req.Host, ":") {
		d, p, _ := net.SplitHostPort(req.Host)
		domain = d
		port, _ = strconv.Atoi(p)
	} else {
		domain = req.Host
	}

	s.host = &routing.TargetHost{
		Domain: domain,
		Port:   port,
	}
	if req.Method == http.MethodConnect {
		s.isTLS = true
		s.host.Port = 443
		if _, err = s.conn.Write([]byte(fmt.Sprintf("%s 200 Connection Established\r\n\r\n", req.Proto))); err != nil {
			return nil, err
		}
	}
	return s.host, nil
}

func (s *httpSession) String() string {
	return fmt.Sprintf("%s://%s:%d", s.host.Domain, s.host.Port, s.host.Port)
}

func (s *httpSession) GetHost() string {
	if s.host.Domain != "" {
		return s.host.Domain
	}
	return s.host.IP.String()
}

func (s *httpSession) GetPort() int {
	return s.host.Port
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
