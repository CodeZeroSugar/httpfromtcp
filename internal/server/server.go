package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/CodeZeroSugar/internal/request"
	"github.com/CodeZeroSugar/internal/response"
)

type Server struct {
	closed   atomic.Bool
	listener net.Listener
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handler Handler) (*Server, error) {
	address := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create 'tcp' listener on address '%s': %w", address, err)
	}

	server := &Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	s.closed.Swap(true)

	return nil
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("failed to accept connection: %s", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	w := response.NewWriter(conn)
	if err != nil {
		s.handler(w, req)
		return
	}
	s.handler(w, req)
}
