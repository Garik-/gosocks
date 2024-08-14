package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

type Server struct {
	listener   net.Listener
	shutdown   chan struct{}
	connection chan net.Conn
	wg         sync.WaitGroup
}

func (s *Server) Start(ctx context.Context) {
	s.wg.Add(2)
	go s.acceptConnections(ctx)
	go s.handleConnections(ctx)
}

func (s *Server) acceptConnections(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}
			s.connection <- conn
		}
	}
}

func (s *Server) handleConnections(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdown:
			return
		case conn := <-s.connection:
			s.wg.Add(1)

			go func() {
				s.handleConnection(ctx, conn)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) Stop(timeout time.Duration) {
	close(s.shutdown)
	_ = s.listener.Close()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-time.After(timeout):
		slog.Warn("timed out waiting for connections to finish")

		return
	}
}

func NewServer(address string) (*Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on address %s: %w", address, err)
	}

	return &Server{
		listener:   listener,
		shutdown:   make(chan struct{}),
		connection: make(chan net.Conn),
	}, nil
}
