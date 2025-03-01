package proxy

import (
	"context"
	"net"
	"strconv"
	"time"
)

type DirectProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Protocol Protocol
}

func (d *DirectProxy) Connect(ctx context.Context, addr string, port uint16) (*Pd, error) {
	dialer := &net.Dialer{
		Timeout: 30 * time.Second,
	}
	agency, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(addr, strconv.FormatUint(uint64(port), 10)))
	if err != nil {
		return nil, err
	}

	return &Pd{agency}, nil
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
