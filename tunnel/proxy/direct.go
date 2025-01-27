package proxy

import (
	"io"
	"net"
	"siuu/logger"
	"strconv"
)

type DirectProxy struct {
	Type     Type
	Name     string
	Server   string
	Port     uint16
	Protocol Protocol
}

func (d *DirectProxy) ForwardTcp(client *Client) error {

	conn := client.Conn
	defer conn.Close()

	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(client.Host, strconv.FormatUint(uint64(client.Port), 10)))
	if err != nil {
		return err
	}

	agency, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer agency.Close()

	if err = agency.SetKeepAlive(true); err != nil {
		return err
	}

	go func() {
		var e error
		if client.Req != nil {
			_, e = io.Copy(agency, client.Req)
		}
		_, e = io.Copy(agency, conn)
		if e != nil {
			logger.SWarn("<%s> %s", client.Sid, e)
		}
	}()

	if _, err = io.Copy(conn, agency); err != nil {
		logger.SWarn("<%s> %s", client.Sid, err)
	}

	return nil
}

func (d *DirectProxy) ForwardUdp(client *Client) (*UdpPocket, error) {
	return nil, ErrProtocolNotSupported
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
