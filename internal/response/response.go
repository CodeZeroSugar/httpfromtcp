package response

import (
	"fmt"
	"strconv"

	"github.com/CodeZeroSugar/internal/headers"
)

type WriterState int

const (
	StatusLine WriterState = 0
	Headers    WriterState = 1
	Body       WriterState = 2
)

type StatusCode int

const (
	StatusCodeOK                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

type Writer struct {
	StatusCode  StatusCode
	Headers     headers.Headers
	Body        string
	writerState WriterState
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case StatusCodeOK:
		reason = "OK"
	case StatusCodeBadRequest:
		reason = "Bad Request"
	case StatusCodeInternalServerError:
		reason = "Internal Server Error"
	default:
		reason = ""
	}
	line := fmt.Sprintf("HTTP/1.1 %d %s", statusCode, reason)
	_, err := w.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("failed to write status line: %w", err)
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, value := range headers {
		payload := key + ": " + value
		_, err := w.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("failed to write '%s' for headers: %w", payload, err)
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to add blank line before body:%w", err)
	}
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return 0, nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen) + "\r\n"
	h["Connection"] = "close\r\n"
	h["Content-Type"] = "text/plain\r\n"
	return h
}
