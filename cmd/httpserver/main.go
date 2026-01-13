package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/CodeZeroSugar/internal/headers"
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

const binURL = "https://httpbin.org/"

func handleHTTPBinProxy(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	fullURL := binURL + path
	h := headers.NewHeaders()
	h.Set("Transfer-Encoding", "chunked")

	resp, err := http.Get(fullURL)
	if err != nil {
		log.Printf("failed to GET response from httpbin: %s", err)
		return
	}
	defer resp.Body.Close()

	if err = w.WriteStatusLine(response.StatusCode(resp.StatusCode)); err != nil {
		log.Printf("failed to write status line: %s", err)
	}
	if err = w.WriteHeaders(h); err != nil {
		log.Printf("failed to write headers: %s", err)
	}

	buf := make([]byte, 1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				log.Printf("failed to write chunked body: %s", err)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("failed to read body of response: %s", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		log.Printf("something went wrong when chunked body was done: %s", err)
	}
}

func handler(w *response.Writer, req *request.Request) {
	path := req.RequestLine.RequestTarget
	if path == "/" {
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
	if path == "/yourproblem" {
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
	if path == "/myproblem" {
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
	if strings.HasPrefix(path, "/httpbin/") {
		handleHTTPBinProxy(w, req)
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
