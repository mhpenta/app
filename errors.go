package app

import (
	"log/slog"
	"strings"
)

const separator = "; "

type MultiError struct {
	Errors []error
}

func AppendError(err error, errs ...error) error {
	if err == nil && len(errs) == 0 {
		return nil
	}

	if len(errs) == 0 {
		mErr, ok := err.(*MultiError)
		if !ok {
			mErr = NewMultiError(err)
		}
		return mErr
	}

	if err == nil {
		return NewMultiError(errs...)
	}

	mErr, ok := err.(*MultiError)
	if !ok {
		mErr = NewMultiError(err)
	}

	for _, e := range errs {
		mErr.Append(e)
	}

	return mErr
}

func NewMultiError(errs ...error) *MultiError {
	mErr := &MultiError{}
	for _, err := range errs {
		mErr.Errors = append(mErr.Errors, err)
	}
	return mErr
}

func (m *MultiError) Append(err error) {
	if err != nil {
		if m == nil {
			slog.Warn("app.MultiError.Append called on nil receiver")
			return
		}

		if m.Errors == nil {
			m.Errors = make([]error, 0)
		}
		m.Errors = append(m.Errors, err)
	}
}

func (m *MultiError) Error() string {
	if m == nil || m.Errors == nil {
		return ""
	}

	if len(m.Errors) == 0 {
		return ""
	}

	if len(m.Errors) < 5 {
		result := m.Errors[0].Error()
		for i := 1; i < len(m.Errors); i++ {
			result += separator + m.Errors[i].Error()
		}
		return result
	} else {
		sb := strings.Builder{}
		sb.WriteString(m.Errors[0].Error())
		for i := 1; i < len(m.Errors); i++ {
			sb.WriteString(separator)
			sb.WriteString(m.Errors[i].Error())
		}
		return sb.String()
	}
}

// ErrorOrNil returns nil if there are no Errors, or the error interface if there are
func (m *MultiError) ErrorOrNil() error {
	if m == nil {
		return nil
	}

	if len(m.Errors) == 0 {
		return nil
	}
	return m
}

func (m *MultiError) HasErrors() bool {
	if m == nil {
		return false
	}
	return len(m.Errors) > 0
}

func (m *MultiError) Unwrap() []error {
	if len(m.Errors) == 0 {
		return nil
	}

	return m.Errors
}
