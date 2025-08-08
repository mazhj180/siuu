package smux

import (
	"errors"
	"io"
	"net"
	"siuu/pkg/proxy/mux"

	"github.com/xtaci/smux/v2"
)

func init() {
	mux.Register("smux", func() mux.Interface {
		return &instance{}
	})
}

type instance struct{}

func (ins *instance) Name() string {
	return "smux"
}

func (ins *instance) Client(conn net.Conn) (mux.Session, error) {
	session, err := smux.Client(conn, smux.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &Session{session}, nil
}

func (ins *instance) Server() (mux.Session, error) {
	return nil, nil
}

type Session struct {
	*smux.Session
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
	*smux.Stream
}

func (s *Stream) CloseWriter() error {
	return s.Stream.Close()
}

func (s *Stream) CloseReader() error {
	return s.Stream.Close()
}
