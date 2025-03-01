package proxy

import (
	"net"
	"strconv"
)

type DirectProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Protocol Protocol
}

func (d *DirectProxy) Connect(addr string, port uint16) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(addr, strconv.FormatUint(uint64(port), 10)))
	if err != nil {
		return nil, err
	}

	agency, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	if err = agency.SetKeepAlive(true); err != nil {
		return nil, err
	}
	return agency, nil
}

func (d *DirectProxy) GetName() string {
	return d.Name
}

func (d *DirectProxy) GetType() Type {
	return d.Type
}

func (d *DirectProxy) GetServer() string {
	return d.Server
}

func (d *DirectProxy) GetPort() uint16 {
	return d.Port
}

func (d *DirectProxy) GetProtocol() Protocol {
	return d.Protocol
}
