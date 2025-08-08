package yamux

import (
	"errors"
	"io"
	"net"
	"siuu/pkg/proxy/mux"

	"github.com/hashicorp/yamux"
)

func init() {
	mux.Register("yamux", func() mux.Interface {
		return &instance{}
	})
}

type instance struct{}

func (ins *instance) Name() string {
	return "yamux"
}

func (ins *instance) Client(conn net.Conn) (mux.Session, error) {
	session, err := yamux.Client(conn, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Session{session}, nil
}

func (ins *instance) Server() (mux.Session, error) {
	return nil, nil
}

type Session struct {
	*yamux.Session
}

func (s *Session) OpenStream() (mux.Stream, error) {
	stream, err := s.Session.OpenStream()
	if err != nil && (errors.Is(err, io.ErrClosedPipe) || errors.Is(err, io.EOF)) {
		return nil, mux.FataErr(err)
	}
	return &Stream{stream}, nil
}
func (s *Session) AcceptStream() (mux.Stream, error) {
	return nil, nil
}

type Stream struct {
	*yamux.Stream
}

func (s *Stream) CloseWriter() error {
	return s.Stream.Close()
}

func (s *Stream) CloseReader() error {
	return s.Stream.Close()
}
