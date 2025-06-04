package proxy

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

type DirectProxy struct{}

func (d *DirectProxy) Type() Type {
	return DIRECT
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

func (d *DirectProxy) Name() string {
	return "direct"
}

func (d *DirectProxy) GetServer() string {
	return ""
}

func (d *DirectProxy) GetPort() uint16 {
	return 0
}

func (d *DirectProxy) GetProtocol() Protocol {
	return TCP
}

func (d *DirectProxy) String() string {
	return fmt.Sprintf(`{"Server":"","Port":0,"Protocol":"","Name":"direct","Type":"%s"}`, d.Type())
}
