package app

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

// TestMetaErrorBasic tests the basic functionality of MetaError.
func TestMetaErrorBasic(t *testing.T) {
	// Create a base error.
	baseErr := errors.New("base error")

	// Wrap the base error using MetaError.
	err := NewMetaError(baseErr)

	// Test that the error message is not empty.
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	// Test that Errors.Is identifies the base error.
	if !errors.Is(err, baseErr) {
		t.Error("Expected Errors.Is to return true for baseErr")
	}

	// Test that the stack trace is captured.
	if err.StackTrace() == "" {
		t.Error("Expected non-empty stack trace")
	}
}

// TestMetaErrorUnwrap tests the Unwrap functionality of MetaError.
func TestMetaErrorUnwrap(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewMetaError(baseErr)

	// Unwrap the error and verify it's the base error.
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != baseErr {
		t.Error("Expected Unwrap to return baseErr")
	}
}

// TestMetaErrorIs tests the Is method with different error types.
func TestMetaErrorIs(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped error: %w", baseErr)
	err := NewMetaError(wrappedErr)

	// Test that Errors.Is identifies baseErr in the error chain.
	if !errors.Is(err, baseErr) {
		t.Error("Expected Errors.Is to return true for baseErr in the chain")
	}

	// Test that Errors.Is identifies wrappedErr.
	if !errors.Is(err, wrappedErr) {
		t.Error("Expected Errors.Is to return true for wrappedErr")
	}
}

// customError is a user-defined error type for testing Errors.As.
type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

// TestMetaErrorAs tests the As method for type assertions.
func TestMetaErrorAs(t *testing.T) {
	baseErr := &customError{"custom error"}
	err := NewMetaError(baseErr)

	var targetErr *customError
	if !errors.As(err, &targetErr) {
		t.Error("Expected Errors.As to successfully assert *customError")
	}

	if targetErr != baseErr {
		t.Error("Expected targetErr to be the same as baseErr")
	}
}

// TestMetaErrorNilError tests MetaError behavior when the underlying error is nil.
func TestMetaErrorNilError(t *testing.T) {
	err := NewMetaError(nil)

	// Test that the error message indicates a nil error.
	if err.Error() != "<nil>" {
		t.Errorf("Expected error message to be '<nil>', got '%s'", err.Error())
	}

	// Test that Unwrap returns nil.
	if errors.Unwrap(err) != nil {
		t.Error("Expected Unwrap to return nil")
	}

	// Stack trace should still be available.
	if err.StackTrace() == "" {
		t.Error("Expected non-empty stack trace even when error is nil")
	}
}

// TestMetaErrorNoStackTrace tests MetaError when the stack trace is not captured.
func TestMetaErrorNoStackTrace(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewMetaErrorOptions(baseErr, 2, false, true)

	// Test that the stack trace is empty.
	if err.StackTrace() != "" {
		t.Error("Expected empty stack trace when captureStack is false")
	}
}

// TestMetaErrorRootCause tests the RootCause function.
func TestMetaErrorRootCause(t *testing.T) {
	baseErr := errors.New("base error")
	wrappedErr := fmt.Errorf("wrapped error: %w", baseErr)
	err := NewMetaError(wrappedErr)

	// Test that RootCause returns the base error.
	root := RootCause(err)
	if root != baseErr {
		t.Error("Expected RootCause to return baseErr")
	}
}

// TestMetaErrorConcurrency tests MetaError in a concurrent context.
func TestMetaErrorConcurrency(t *testing.T) {
	baseErr := errors.New("base error")
	const numGoroutines = 10
	errorsCh := make(chan error, numGoroutines)

	// Run multiple goroutines to test thread safety.
	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := NewMetaError(baseErr)
			errorsCh <- err
		}()
	}

	// Collect and verify Errors from all goroutines.
	for i := 0; i < numGoroutines; i++ {
		err := <-errorsCh
		if err.Error() == "" {
			t.Error("Expected non-empty error message in goroutine")
		}
	}
}

// TestMetaErrorFormattingWithErrorf tests formatting with fmt.Errorf.
func TestMetaErrorFormattingWithErrorf(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewMetaError(baseErr)

	// Use fmt.Errorf with MetaError.
	formattedErr := fmt.Errorf("an error occurred: %w", err)

	// Test that Errors.Is identifies baseErr in the chain.
	if !errors.Is(formattedErr, baseErr) {
		t.Error("Expected Errors.Is to return true for baseErr in formattedErr")
	}

	// Test that the formatted error message includes the MetaError message.
	if !strings.Contains(formattedErr.Error(), err.Error()) {
		t.Error("Expected formatted error message to include MetaError message")
	}
}

func TestMetaError(t *testing.T) {
	baseErr := errors.New("base error")
	err := NewMetaError(baseErr)

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	if !errors.Is(err, baseErr) {
		t.Error("Expected Errors.Is to return true")
	}

	if err.StackTrace() == "" {
		t.Error("Expected non-empty stack trace")
	}
}

func TestAnonymousFunc(t *testing.T) {
	result := isAnonymousFuncName("modeledge-go/internal/routes.RegisterToolRoutes.FakeError")
	if result {
		t.Error("Expected false, received true")
	}

	result = isAnonymousFuncName("modeledge-go/internal/routes.RegisterToolRoutes.FakeError.func1")
	if !result {
		t.Error("Expected true, received false")
	}
}

func TestMetaErrorFuncName(t *testing.T) {
	var err *MetaError
	func() {
		baseErr := errors.New("base error")
		err = NewMetaError(baseErr)
	}()

	if err.Func == "" {
		t.Error("Expected non-empty function name")
	}

	if err.Package == "" {
		t.Error("Expected non-empty package name")
	}

	if err.File == "" {
		t.Error("Expected non-empty file name")
	}

	if err.Func != "TestMetaErrorFuncName" {
		t.Error("Expected function name to be TestMetaErrorFuncName, received: ", err.Func)
	}

	if err.Package != "modeledge-go/ext/app" {
		t.Error("Expected function name to be TestMetaErrorFuncName, received: ", err.Func)
	}

	pkgPath, recvName, recvPtr, typeGeneric, funcGeneric, funcNameSmall, notice := parseFuncName("modeledge-go/internal/routes.RegisterToolRoutes.FakeError.func1")
	slog.Info("funcname", "pkgPath", pkgPath, "recvName", recvName, "recvPtr", recvPtr, "typeGeneric", typeGeneric, "funcGeneric", funcGeneric, "funcNameSmall", funcNameSmall)
	slog.Info("funcname", "notice", notice)

	pkgPath, recvName, recvPtr, typeGeneric, funcGeneric, funcNameSmall, notice = parseFuncName("modeledge-go/internal/routes.RegisterToolRoutes.FakeError.func1")
	slog.Info("funcname", "pkgPath", pkgPath, "recvName", recvName, "recvPtr", recvPtr, "typeGeneric", typeGeneric, "funcGeneric", funcGeneric, "funcNameSmall", funcNameSmall)
	slog.Info("funcname", "notice", notice)

	/*
		//  TODO fix
		if err.Func != "TestMetaErrorFuncName" {
			t.Error("Expected function name to be TestMetaErrorFuncName, received: ", err.Func)
		}

		if err.Package != "modeledge-go/ext/app" {
			t.Error("Expected function name to be \"modeledge-go/ext/app, received: ", err.Package)
		}*/
}
