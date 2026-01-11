package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/CodeZeroSugar/internal/request"
	"github.com/CodeZeroSugar/internal/response"
)

type Server struct {
	ServerState *atomic.Bool
	Listener    net.Listener
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteError(conn io.Writer, error *HandlerError) error {
	payload := fmt.Sprintf("%d %s\r\n", error.StatusCode, error.Message)
	_, err := conn.Write([]byte(payload))
	if err != nil {
		return fmt.Errorf("failed to write HandlerError to connection: %w", err)
	}
	return nil
}

func Serve(port int, handler Handler) (*Server, error) {
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

	go server.listen(handler)

	return &server, nil
}

func (s *Server) Close() error {
	if err := s.Listener.Close(); err != nil {
		return fmt.Errorf("failed to close listener: %w", err)
	}
	s.ServerState.Swap(false)

	return nil
}

func (s *Server) listen(handler Handler) {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if !s.ServerState.Load() {
				return
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
			log.Printf("failed to accept connection: %s", err)
			continue
		}

		go s.handle(conn, handler)
	}
}

func (s *Server) handle(conn net.Conn, handler Handler) {
	defer conn.Close()
	request, err := request.RequestFromReader(conn)
	if err != nil {
		log.Printf("failed to read request from connection: %s", err)
	}
	buf := new(bytes.Buffer)
	handlerError := handler(buf, request)
	if handlerError != nil {
		err = WriteError(conn, handlerError)
		if err != nil {
			log.Printf("failed to write error to connection: %s", err)
		}
	} else {
		h := response.GetDefaultHeaders(0)
		err := response.WriteStatusLine(conn, response.OK)
		if err != nil {
			log.Printf("failed to write status line: %s", err)
		}
		err = response.WriteHeaders(conn, h)
		if err != nil {
			log.Printf("failed to write headers: %s", err)
		}
	}
}
