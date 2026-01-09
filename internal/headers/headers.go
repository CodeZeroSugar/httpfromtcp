package headers

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var validFieldName = regexp.MustCompile(`^[A-Za-z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)

func validateFieldName(fieldName string) error {
	if !validFieldName.MatchString(fieldName) {
		return errors.New("malformed data, invalid character found in field name")
	}
	return nil
}

type Headers map[string]string

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	if bytes.Index(data, []byte("\r\n")) == 0 {
		return 2, true, nil
	}

	str := string(data)

	splitNewLine := strings.Split(str, "\r\n")
	if len(splitNewLine) == 1 {
		return 0, false, nil
	}

	headerLine := splitNewLine[0]

	splitColon := strings.Split(headerLine, ":")
	if len(splitColon) == 1 {
		return 0, false, errors.New("malformed data, could not find ':'")
	}

	n += len(headerLine) + 2

	fieldName := strings.ToLower(strings.TrimLeft(splitColon[0], " "))
	if len(fieldName) < 1 {
		return 0, false, errors.New("field name is null")
	}

	if err = validateFieldName(fieldName); err != nil {
		return 0, false, err
	}

	runes := []rune(fieldName)
	if unicode.IsSpace(runes[len(runes)-1]) {
		return 0, false, errors.New("invalid whitespace found before ':' character")
	}

	trimmedValues := make([]string, 0)

	for _, slice := range splitColon[1:] {
		trimmed := strings.Trim(slice, " ")
		trimmedValues = append(trimmedValues, trimmed)
	}

	fieldValue := strings.Join(trimmedValues, ":")
	_, exists := h[fieldName]
	if exists {
		h[fieldName] += ", " + fieldValue
	} else {
		h[fieldName] = fieldValue
	}

	return n, false, nil
}
