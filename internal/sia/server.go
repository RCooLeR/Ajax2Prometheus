package sia

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/rs/zerolog"
)

type Handler func(ctx context.Context, raw []byte, remoteAddr string) ([]byte, error)

type Server struct {
	addr        string
	readTimeout time.Duration
	handler     Handler
	log         zerolog.Logger
}

func NewServer(addr string, readTimeout time.Duration, handler Handler, log zerolog.Logger) *Server {
	return &Server{addr: addr, readTimeout: readTimeout, handler: handler, log: log}
}

func (s *Server) Run(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.log.Info().Str("addr", listener.Addr().String()).Msg("SIA listener started")
	go func() {
		<-ctx.Done()
		_ = listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return nil
			}
			s.log.Error().Err(err).Msg("accept SIA connection")
			continue
		}
		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log := s.log.With().Str("remote", remote).Logger()
	log.Debug().Msg("SIA connection opened")
	defer log.Debug().Msg("SIA connection closed")

	reader := bufio.NewReader(conn)
	for {
		if s.readTimeout > 0 {
			_ = conn.SetReadDeadline(time.Now().Add(s.readTimeout))
		}
		frame, err := reader.ReadBytes('\r')
		if len(frame) > 0 {
			response, handleErr := s.handler(ctx, frame, remote)
			if handleErr != nil {
				log.Warn().Err(handleErr).Msg("SIA frame handled with error")
			}
			if len(response) > 0 {
				if _, writeErr := conn.Write(response); writeErr != nil {
					log.Warn().Err(writeErr).Msg("write SIA response")
					return
				}
			}
		}
		if err == nil {
			continue
		}
		if errors.Is(err, io.EOF) {
			return
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return
		}
		log.Warn().Err(fmt.Errorf("read SIA frame: %w", err)).Msg("read SIA frame")
		return
	}
}
