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
	closed   *atomic.Bool
	listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func WriteError(conn io.Writer, herr *HandlerError) error {
	err := response.WriteStatusLine(conn, herr.StatusCode)
	if err != nil {
		return fmt.Errorf("failed to write status line in write error: %w", err)
	}
	h := response.GetDefaultHeaders(len(herr.Message))
	err = response.WriteHeaders(conn, h)
	if err != nil {
		return fmt.Errorf("failed to write headers in write error: %w", err)
	}
	_, err = conn.Write([]byte(herr.Message))
	if err != nil {
		return fmt.Errorf("failed to write message for handler error: %w", err)
	}
	return nil
}

func Serve(port int, handler Handler) (*Server, error) {
	address := ":" + strconv.Itoa(port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to create 'tcp' listener on address '%s': %w", address, err)
	}

	server := Server{
		listener: listener,
		handler:  handler,
	}

	go server.listen()

	return &server, nil
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
	request, err := request.RequestFromReader(conn)
	if err != nil {
		herr := HandlerError{
			StatusCode: response.StatusCodeBadRequest,
			Message:    "something wasn't handled correctly",
		}
		log.Printf("request from reader failed, sending notification to client: %s", err)
		err = WriteError(conn, &herr)
		if err != nil {
			log.Printf("write error caused an error: %s", err)
		}
		return
	}
	buf := new(bytes.Buffer)
	handlerError := s.handler(buf, request)
	if handlerError != nil {
		err = WriteError(conn, handlerError)
		if err != nil {
			log.Printf("failed to write error to connection: %s", err)
		}
	} else {
		h := response.GetDefaultHeaders(buf.Len())
		err := response.WriteStatusLine(conn, response.StatusCodeOK)
		if err != nil {
			log.Printf("failed to write status line: %s", err)
		}
		err = response.WriteHeaders(conn, h)
		if err != nil {
			log.Printf("failed to write headers: %s", err)
		}
		_, err = conn.Write(buf.Bytes())
		if err != nil {
			log.Printf("failed to write body: %s", err)
		}
	}
}
