package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func parseRequestLine(buff []byte) (RequestLine, error) {
	str := string(buff)
	lines := strings.Split(str, "\r\n")
	split := strings.Split(lines[0], " ")

	if len(split) != 3 {
		return RequestLine{}, fmt.Errorf("invalid number of parts in request line")
	}

	m, t, v := split[0], split[1], split[2]

	for _, r := range m {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return RequestLine{}, fmt.Errorf("found a non-alphabet or lowercase character while parsing request line")
		}
	}

	splitVersion := strings.Split(v, "/")
	if splitVersion[1] != "1.1" {
		return RequestLine{}, fmt.Errorf("http version was not '1.1'")
	}
	ver := splitVersion[1]

	return RequestLine{
		HttpVersion:   ver,
		RequestTarget: t,
		Method:        m,
	}, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buff, err := io.ReadAll(reader)
	if err != nil {
		return &Request{}, fmt.Errorf("failed to read into bytes: %w", err)
	}

	req, err := parseRequestLine(buff)
	if err != nil {
		return &Request{}, fmt.Errorf("failed to parse request line: %w", err)
	}

	return &Request{RequestLine: req}, nil
}
