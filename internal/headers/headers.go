package headers

import (
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)
	n = 0

	hasNewLine := strings.Contains(str, "\r\n")
	if !hasNewLine {
		return 0, false, nil
	}

	for i := 0; i < len(str)-2; i++ {
		if str[i:i+2] == "\r\n" {
			n = i + 2
			break
		}
	}
	if n == 0 {
		return n, true, nil
	}

	field := str[:n-2]
	splitField := strings.Split(field, ":")
	fieldName := strings.TrimLeft(splitField[0], " ")

	runes := []rune(fieldName)
	if unicode.IsSpace(runes[len(runes)-1]) {
		return 0, false, fmt.Errorf("invalid whitespace found before ':' character: %w", err)
	}

	fieldValue := strings.Trim(splitField[1], " ") + ":" + strings.Trim(splitField[2], " ")

	h[fieldName] = fieldValue
	return n, false, nil
}
