package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CodeZeroSugar/internal/request"
	"github.com/CodeZeroSugar/internal/response"
	"github.com/CodeZeroSugar/internal/server"
)

const port = 42069

func handler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		notFound := server.HandlerError{
			StatusCode: response.BadRequest,
			Message:    "Your problem is not my problem\n",
		}
		return &notFound
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		serverError := server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "Woopsie, my bad\n",
		}
		return &serverError
	}
	msg := "All good, frfr\n"
	_, err := w.Write([]byte(msg))
	if err != nil {
		serverError := server.HandlerError{
			StatusCode: response.InternalServerError,
			Message:    "Woopsie, my bad\n",
		}
		return &serverError
	}
	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
