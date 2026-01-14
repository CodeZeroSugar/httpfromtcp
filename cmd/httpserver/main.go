package main

import (
	"crypto/sha256"
	"fmt"
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

const (
	binURL   = "https://httpbin.org/"
	xContent = "X-Content-Sha256"
	xLength  = "X-Content-Length"
)

func handleHTTPBinProxy(w *response.Writer, req *request.Request) {
	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
	fullURL := binURL + path
	h := response.GetDefaultHeaders(0)
	h.Set("Transfer-Encoding", "chunked")
	h.Set("Trailer", xContent+", "+xLength)
	h.Del("Content-length")

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
	var bodyBuf []byte

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
			bodyBuf = append(bodyBuf, buf[:n]...)
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
	trailers := headers.NewHeaders()

	shaHash := sha256.Sum256(bodyBuf)
	trailers.Set(xContent, fmt.Sprintf("%x", shaHash))
	trailers.Set(xLength, fmt.Sprintf("%d", len(bodyBuf)))
	if err = w.WriteTrailers(trailers); err != nil {
		log.Printf("failed to write trailers: %s", err)
		return
	}
}

func handleVideo(w *response.Writer, req *request.Request) {
	file, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		log.Printf("failed to read video: %s", err)
	}
	h := response.GetDefaultHeaders(len(file))
	h.Set("Content-Type", "video/mp4")
	if err := w.WriteStatusLine(response.StatusCodeOK); err != nil {
		log.Printf("failed to write status line for video: %s", err)
		return
	}
	if err := w.WriteHeaders(h); err != nil {
		log.Printf("failed to write headers for video: %s", err)
		return
	}
	if _, err := w.WriteBody(file); err != nil {
		log.Printf("failed to write video: %s", err)
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
	if path == "/video" {
		handleVideo(w, req)
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
