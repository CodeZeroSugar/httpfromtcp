package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CodeZeroSugar/internal/request"
	"github.com/CodeZeroSugar/internal/response"
	"github.com/CodeZeroSugar/internal/server"
)

const okHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

const badRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrorHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const (
	port = 42069
)

func handler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/" {
		body := []byte(okHTML)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		if err := w.WriteStatusLine(response.StatusCodeOK); err != nil {
			log.Printf("handler failed to write status line: %s", err)
		}
		if err := w.WriteHeaders(h); err != nil {
			log.Printf("handler failed to write headers: %s", err)
		}
		if _, err := w.WriteBody(body); err != nil {
			log.Printf("handler failed to write body: %s", err)
		}
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		body := []byte(badRequestHTML)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		if err := w.WriteStatusLine(response.StatusCodeBadRequest); err != nil {
			log.Printf("handler failed to write status line: %s", err)
		}
		if err := w.WriteHeaders(h); err != nil {
			log.Printf("handler failed to write headers: %s", err)
		}
		if _, err := w.WriteBody(body); err != nil {
			log.Printf("handler failed to write body: %s", err)
		}
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		body := []byte(internalErrorHTML)
		h := response.GetDefaultHeaders(len(body))
		h["Content-Type"] = "text/html"
		if err := w.WriteStatusLine(response.StatusCodeInternalServerError); err != nil {
			log.Printf("handler failed to write status line: %s", err)
		}
		if err := w.WriteHeaders(h); err != nil {
			log.Printf("handler failed to write headers: %s", err)
		}
		if _, err := w.WriteBody(body); err != nil {
			log.Printf("handler failed to write body: %s", err)
		}
	}
}

func main() {
	srv, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer srv.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
