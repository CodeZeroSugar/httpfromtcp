package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/CodeZeroSugar/internal/headers"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	ParserState ParserState
}

type ParserState int

const (
	requestStateInitialized    = 0
	requestStateDone           = 1
	requestStateParsingHeaders = 2
	requestStateParsingBody    = 3
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.ParserState != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.ParserState {
	case requestStateInitialized:
		line, n, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("failed to parse request line: %w", err)
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = line
		r.ParserState = requestStateParsingHeaders

		return n, nil

	case requestStateParsingHeaders:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.ParserState = requestStateParsingBody
		}
		return n, nil

	case requestStateParsingBody:
		value, exists := r.Headers.Get("Content-Length")
		if !exists {
			r.ParserState = requestStateDone
			return 0, nil
		}
		contentLength, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("content length failed to convert to int: %w", err)
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLength {
			return 0, fmt.Errorf("length of data was greater than Content-Length %d:%d", len(r.Body), contentLength)
		}
		if len(r.Body) == contentLength {
			r.ParserState = requestStateDone
		}

		return len(r.Body), nil

	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")

	default:
		return 0, errors.New("error: unknown state")
	}
}

func parseRequestLine(buff []byte) (RequestLine, int, error) {
	str := string(buff)

	hasNewLine := strings.Contains(str, "\r\n")
	if !hasNewLine {
		return RequestLine{}, 0, nil
	}
	parts := strings.SplitAfter(str, "\r\n")
	firstLineWithCRLF := parts[0]
	n := len(firstLineWithCRLF)

	linesText := firstLineWithCRLF[:n-2]

	fields := strings.Split(linesText, " ")

	if len(fields) != 3 {
		return RequestLine{}, n, fmt.Errorf("invalid number of parts in request line")
	}

	m, t, v := fields[0], fields[1], fields[2]

	for _, r := range m {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return RequestLine{}, n, fmt.Errorf("found a non-alphabet or lowercase character while parsing request line")
		}
	}

	splitVersion := strings.Split(v, "/")
	if splitVersion[1] != "1.1" {
		return RequestLine{}, n, fmt.Errorf("http version was not '1.1'")
	}
	ver := splitVersion[1]

	return RequestLine{
		HttpVersion:   ver,
		RequestTarget: t,
		Method:        m,
	}, n, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff := make([]byte, bufferSize)

	var readToIndex int
	readToIndex = 0

	req := Request{
		ParserState: requestStateInitialized,
		Headers:     headers.NewHeaders(),
	}
	for req.ParserState != requestStateDone {
		if readToIndex >= len(buff) {
			newBuff := make([]byte, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}
		n, err := reader.Read(buff[readToIndex:])
		if err == io.EOF {
			if req.ParserState != requestStateDone {
				return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.ParserState, n)
			}
			break
		}
		if err != nil {
			return &Request{}, fmt.Errorf("failed to read from index: %w", err)
		}
		readToIndex += n

		bytesConsumed, err := req.parse(buff[:readToIndex])
		if err != nil {
			return &Request{}, fmt.Errorf("failed to parse request: %w", err)
		}

		copy(buff, buff[bytesConsumed:readToIndex])

		readToIndex -= bytesConsumed

	}
	return &req, nil
}
