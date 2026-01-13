package response

import (
	"fmt"
	"io"
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
	conn        io.Writer
	writerState WriterState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		conn:        w,
		writerState: StatusLine,
	}
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	hexString := fmt.Sprintf("%02X\r\n", len(p))
	hexBytes := []byte(hexString)
	hexBytes = append(hexBytes, p...)
	hexBytes = append(hexBytes, "\r\n"...)
	_, err := w.WriteBody(hexBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to write chunked body: %w", err)
	}
	return len(p), nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	chunkDone := "0\r\n\r\n"
	n, err := w.WriteBody([]byte(chunkDone))
	if err != nil {
		return 0, fmt.Errorf("failed to write chunked body as done: %w", err)
	}
	return n, nil
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState == StatusLine {
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
		line := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
		_, err := w.conn.Write([]byte(line))
		if err != nil {
			return fmt.Errorf("failed to write status line: %w", err)
		}
		w.writerState = Headers
		return nil
	}
	return fmt.Errorf("tried to write status line while state was: %v", w.writerState)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState == Headers {
		for key, value := range headers {
			payload := key + ": " + value + "\r\n"
			_, err := w.conn.Write([]byte(payload))
			if err != nil {
				return fmt.Errorf("failed to write '%s' for headers: %w", payload, err)
			}
		}
		_, err := w.conn.Write([]byte("\r\n"))
		if err != nil {
			return fmt.Errorf("failed to add blank line before body:%w", err)
		}
		w.writerState = Body
		return nil
	}
	return fmt.Errorf("tried to write headers while state was: %v", w.writerState)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState == Body {
		n, err := w.conn.Write(p)
		if err != nil {
			return 0, fmt.Errorf("failed to write body: %w", err)
		}
		return n, nil
	}
	return 0, fmt.Errorf("tried to write body while state was: %v", w.writerState)
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}
