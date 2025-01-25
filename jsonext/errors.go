package jsonext

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

func IsUnmarshallingError(err error) bool {
	if err == nil {
		return false
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return true
	}

	var unmarshalTypeErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalTypeErr) {
		return true
	}

	var invalidUnmarshalErr *json.InvalidUnmarshalError
	if errors.As(err, &invalidUnmarshalErr) {
		return true
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	errStr := err.Error()
	commonErrors := []string{
		"invalid character",
		"cannot unmarshal",
		"unexpected end of JSON input",
	}

	for _, phrase := range commonErrors {
		if strings.Contains(errStr, phrase) {
			return true
		}
	}

	return false
}
