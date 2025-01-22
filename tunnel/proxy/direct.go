package proxy

import (
	"errors"
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

func (d *DirectProxy) Act(client *Client) error {
	if d.Protocol == TCP {
		if err := d.actOfTcp(client); err != nil {
			return err
		}
	} else if d.Protocol == UDP {
		if err := d.actOfUdp(client); err != nil {
			return err
		}
	} else {
		return ErrProtocolNotSupported
	}
	return nil
}

func (d *DirectProxy) actOfTcp(client *Client) error {

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

func (d *DirectProxy) actOfUdp(client *Client) error {
	return errors.New("it is not supported udp yet")
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
