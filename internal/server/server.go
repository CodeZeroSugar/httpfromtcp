package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/CodeZeroSugar/internal/response"
)

type Server struct {
	ServerState *atomic.Bool
	Listener    net.Listener
}

func Serve(port int) (*Server, error) {
	address := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return &Server{}, fmt.Errorf("failed to create 'tcp' listener on address '%s': %w", address, err)
	}
	state := atomic.Bool{}
	state.Store(true)

	server := Server{
		ServerState: &state,
		Listener:    listener,
	}

	go server.listen()

	return &server, nil
}

func (s *Server) Close() error {
	if err := s.Listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	s.ServerState.Swap(false)

	return nil
}

func (s *Server) listen() {
	for {
		if !s.ServerState.Load() {
			return
		}
		conn, err := s.Listener.Accept()
		if err != nil {
			if !s.ServerState.Load() {
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
	resp := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Hello World!\n"

	err := response.WriteStatusLine(conn, 200)
	if err != nil {
		log.Printf("failed to write status line: %s", err)
	}
	err = response.WriteHeaders(conn)

	_, err := conn.Write([]byte(resp))
	if err != nil {
		log.Printf("failed to write response to connection: %s", err)
	}
}
