package response

import (
	"fmt"
	"io"
	"strconv"

	"github.com/CodeZeroSugar/internal/headers"
)

type StatusCode int

const (
	OK                  = 200
	BadRequest          = 400
	InternalServerError = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case OK:
	case BadRequest:
		reason = "HTTP/1.1 400 Bad Request"
	case InternalServerError:
		reason = "HTTP/1.1 500 Internal Server Error"
	default:
		reason = ""
	}
	_, err := w.Write([]byte(reason))
	if err != nil {
		return fmt.Errorf("failed to write reason for status code '%d': %w", statusCode, err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		payload := key + ": " + value
		_, err := w.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("failed to write '%s' for headers: %w", payload, err)
		}
	}
	return nil
}
