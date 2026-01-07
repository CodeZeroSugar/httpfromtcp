package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
}

type ParserState int

const (
	initialized = 0
	done        = 1
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == initialized {
		line, n, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("failed to parse request line: %w", err)
		}
		if n == 0 {
			return 0, nil
		}

		r.RequestLine = line
		r.ParserState = done

		return n, nil

	} else if r.ParserState == done {
		return 0, errors.New("error: trying to read data in a done state")
	} else {
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
		ParserState: initialized,
	}
	for req.ParserState != done {
		if readToIndex >= len(buff) {
			newBuff := make([]byte, len(buff)*2)
			copy(newBuff, buff)
			buff = newBuff
		}
		n, err := reader.Read(buff[readToIndex:])
		if err == io.EOF {
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
