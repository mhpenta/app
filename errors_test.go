package app

import (
	"errors"
	"fmt"
	"testing"
)

func TestMultiError_Append(t *testing.T) {
	tests := []struct {
		name    string
		errors  []error
		wantStr string
		wantNil bool
	}{
		{
			name:    "no Errors, 1",
			errors:  []error{nil},
			wantStr: "",
			wantNil: true,
		},
		{
			name:    "no Errors, 2",
			errors:  []error{},
			wantStr: "",
			wantNil: true,
		},
		{
			name:    "no Errors, 3",
			errors:  nil,
			wantStr: "",
			wantNil: true,
		},
		{
			name:    "no Errors, 4",
			errors:  make([]error, 0, 10),
			wantStr: "",
			wantNil: true,
		},
		{
			name:    "single error",
			errors:  []error{errors.New("error one")},
			wantStr: "error one",
			wantNil: false,
		},
		{
			name:    "multiple Errors",
			errors:  []error{errors.New("error one"), errors.New("error two")},
			wantStr: "error one; error two",
			wantNil: false,
		},
		{
			name:    "mixed nil and Errors",
			errors:  []error{nil, errors.New("error one"), nil, errors.New("error two")},
			wantStr: "error one; error two",
			wantNil: false,
		},
		{
			name:    "all nil Errors",
			errors:  []error{nil, nil, nil},
			wantStr: "",
			wantNil: true,
		},
		{
			name: "more than five Errors",
			errors: []error{
				errors.New("error one"),
				errors.New("error two"),
				errors.New("error three"),
				errors.New("broken pipe"),
				errors.New("error five"),
				errors.New("error six"),
			},
			wantStr: "error one; error two; error three; broken pipe; error five; error six",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m MultiError

			// Append all Errors in the test case
			for _, err := range tt.errors {
				m.Append(err)
			}

			// Test Error() string output
			if got := m.Error(); got != tt.wantStr {
				t.Errorf("MultiError.Error() = %v, want %v", got, tt.wantStr)
			}

			// Test ErrorOrNil() behavior
			if got := m.ErrorOrNil(); (got == nil) != tt.wantNil {
				t.Errorf("MultiError.ErrorOrNil() = %v, want nil: %v", got, tt.wantNil)
			}

			// Test HasErrors() behavior
			if got := m.HasErrors(); got == tt.wantNil {
				t.Errorf("MultiError.HasErrors() = %v, want %v", got, !tt.wantNil)
			}
		})
	}
}

func TestMultiError_AppendNil(t *testing.T) {
	var m MultiError
	m.Append(nil)

	if m.HasErrors() {
		t.Error("MultiError should not have Errors after appending nil")
	}

	if err := m.ErrorOrNil(); err != nil {
		t.Error("MultiError.ErrorOrNil() should return nil when no Errors are present")
	}
}

func TestMultiError_Order(t *testing.T) {
	var m MultiError

	err1 := errors.New("first")
	err2 := errors.New("second")
	err3 := errors.New("third")

	m.Append(err1)
	m.Append(err2)
	m.Append(err3)

	expected := "first; second; third"
	if got := m.Error(); got != expected {
		t.Errorf("MultiError.Error() = %v, want %v", got, expected)
	}
}

func TestMultiError_EmptyBehavior(t *testing.T) {
	var m MultiError

	if m.HasErrors() {
		t.Error("Empty MultiError should not have Errors")
	}

	if got := m.Error(); got != "" {
		t.Errorf("Empty MultiError.Error() = %v, want empty string", got)
	}

	if err := m.ErrorOrNil(); err != nil {
		t.Error("Empty MultiError.ErrorOrNil() should return nil")
	}
}

func TestMultiError_ErrorInterface(t *testing.T) {
	var m MultiError
	var err error = &m // Verify it implements error interface

	m.Append(errors.New("test error"))

	if err.Error() != "test error" {
		t.Errorf("Error interface implementation failed, got %v", err.Error())
	}
}

func TestMultiError_StringsBuilder(t *testing.T) {
	var m MultiError

	// Create more than 5 Errors to trigger strings.Builder path
	errors := []error{
		errors.New("error one"),
		errors.New("error two"),
		errors.New("error three"),
		errors.New("error four"),
		errors.New("error five"),
		errors.New("error six"),
	}

	for _, err := range errors {
		m.Append(err)
	}

	expected := "error one; error two; error three; error four; error five; error six"
	if got := m.Error(); got != expected {
		t.Errorf("MultiError.Error() = %v, want %v", got, expected)
	}
}

type TestErr struct {
	msg string
}

func (e *TestErr) Error() string {
	return e.msg
}

func TestMultiError_IsAs(t *testing.T) {
	var m MultiError
	var tErr *TestErr

	brokenPipe := errors.New("broken pipe")

	variousErrors := []error{
		errors.New("error one"),
		errors.New("error two"),
		errors.New("error three"),
		fmt.Errorf("wrapped: %w", brokenPipe),
		errors.New("error five"),
		errors.New("error six"),
		&TestErr{"test error"},
	}

	for _, err := range variousErrors {
		m.Append(err)
	}

	if errors.Is(&m, nil) {
		t.Error("MultiError should not be nil")
	}

	if !errors.Is(&m, brokenPipe) {
		t.Error("MultiError should contain broken pipe error")
	}

	if errors.Is(&m, errors.New("error one")) {
		t.Error("MultiError should not contain error one")
	}

	if !errors.As(&m, &tErr) {
		t.Error("MultiError should contain TestErr")
	}
}

func TestAppendError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		errs    []error
		wantStr string
		wantNil bool
	}{
		{
			name:    "nil error with no additional errors",
			err:     nil,
			errs:    nil,
			wantStr: "",
			wantNil: true,
		},
		{
			name:    "nil error with single additional error",
			err:     nil,
			errs:    []error{errors.New("error one")},
			wantStr: "error one",
			wantNil: false,
		},
		{
			name:    "single error with no additional errors",
			err:     errors.New("error one"),
			errs:    nil,
			wantStr: "error one",
			wantNil: false,
		},
		{
			name:    "single error with additional error",
			err:     errors.New("error one"),
			errs:    []error{errors.New("error two")},
			wantStr: "error one; error two",
			wantNil: false,
		},
		{
			name:    "MultiError with additional errors",
			err:     NewMultiError(errors.New("error one")),
			errs:    []error{errors.New("error two"), errors.New("error three")},
			wantStr: "error one; error two; error three",
			wantNil: false,
		},
		{
			name:    "error with nil additional errors",
			err:     errors.New("error one"),
			errs:    []error{nil, nil},
			wantStr: "error one",
			wantNil: false,
		},
		{
			name:    "MultiError with mixed nil and non-nil additional errors",
			err:     NewMultiError(errors.New("error one")),
			errs:    []error{nil, errors.New("error two"), nil},
			wantStr: "error one; error two",
			wantNil: false,
		},
		{
			name: "many errors to trigger strings.Builder",
			err:  errors.New("error one"),
			errs: []error{
				errors.New("error two"),
				errors.New("error three"),
				errors.New("error four"),
				errors.New("error five"),
				errors.New("error six"),
			},
			wantStr: "error one; error two; error three; error four; error five; error six",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AppendError(tt.err, tt.errs...)

			if tt.wantNil {
				if result != nil {
					t.Errorf("AppendError() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Error("AppendError() returned nil, want non-nil error")
				return
			}

			if got := result.Error(); got != tt.wantStr {
				t.Errorf("AppendError().Error() = %v, want %v", got, tt.wantStr)
			}

			// Test that the result implements error unwrapping
			if mErr, ok := result.(*MultiError); ok {
				unwrapped := mErr.Unwrap()
				totalExpectedErrors := len(tt.errs)
				if tt.err != nil {
					totalExpectedErrors++
				}

				// Count non-nil errors in tt.errs
				nonNilErrors := 0
				for _, err := range tt.errs {
					if err != nil {
						nonNilErrors++
					}
				}
				if tt.err != nil {
					nonNilErrors++
				}

				if len(unwrapped) != nonNilErrors {
					t.Errorf("AppendError() unwrapped length = %v, want %v", len(unwrapped), nonNilErrors)
				}
			} else {
				t.Error("AppendError() result does not implement error unwrapping")
			}
		})
	}
}
