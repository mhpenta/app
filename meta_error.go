package app

import (
	"encoding/csv"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const initialStackSize = 64
const maxStackDepth = 1024 // To prevent excessive memory usage

var ErrNotMetaError = errors.New("error is not a MetaError")

// MetaError wraps an error with additional context information such as file,
// line number, function name, package name, and stack trace.
type MetaError struct {
	Err              error
	File             string
	Line             int
	Func             string
	Package          string
	stackTrace       []uintptr
	stackTraceString string
	asCSV            bool
}

// Errorf creates a new MetaError with the given format and arguments and captures the stack trace.
func Errorf(format string, args ...interface{}) *MetaError {
	return NewMetaError(fmt.Errorf(format, args...))
}

// NewMetaError creates a new MetaError with the given error and captures the stack trace.
// If the given error is already a *MetaError, it is returned as-is to preserve
// its original context and avoid redundant wrapping. Note that this check
// only works for direct *MetaError types and not for wrapped errors containing a MetaError.
func NewMetaError(err error) *MetaError {
	if metaErr, ok := err.(*MetaError); ok {
		return metaErr
	}
	return NewMetaErrorOptions(err, 2, true, true) // Skip 2 frames
}

func Slog(err error) []interface{} {
	metaError := NewMetaError(err)

	if metaError == nil {
		return []interface{}{}
	}

	return []interface{}{
		"error_meta", err,
		"file_meta", metaError.File,
		"line_meta", metaError.Line,
		"func_meta", metaError.Func,
	}
}

func NewMetaErrorOptions(err error, skip int, captureStack bool, asCSV bool) *MetaError {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "unknown"
		line = 0
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	packageName := "unknown"

	if fn != nil {
		fullFuncName := fn.Name()
		lastDotIndex := strings.LastIndex(fullFuncName, ".")
		if lastDotIndex != -1 {
			packageName = fullFuncName[:lastDotIndex]
			funcName = fullFuncName[lastDotIndex+1:]
		} else {
			funcName = fullFuncName
		}
	}

	metaErr := &MetaError{
		Err:     err,
		File:    filepath.Base(file),
		Line:    line,
		Func:    funcName,
		Package: packageName,
		asCSV:   asCSV,
	}

	if captureStack {
		pcs := make([]uintptr, initialStackSize)
		n := runtime.Callers(skip, pcs)
		for n == len(pcs) && len(pcs) < maxStackDepth {
			pcs = make([]uintptr, len(pcs)*2)
			n = runtime.Callers(skip, pcs)
		}
		if len(pcs) > maxStackDepth {
			pcs = pcs[:maxStackDepth]
		}
		metaErr.stackTrace = pcs[:n]
	}

	return metaErr
}

// Error returns the error message with context.
func (e *MetaError) Error() string {
	if e.Err == nil {
		return "<nil>"
	}
	return e.Err.Error()
}

func (e *MetaError) Format(s fmt.State, verb rune) {
	if e.asCSV {
		switch verb {
		case 'v':
			if s.Flag('+') {
				fmt.Fprintf(s, "%s|%s", e.ToCSV(), e.StackTrace())
			}
			fallthrough
		default:
			fmt.Fprintf(s, "%s", e.ToCSV())
		}
		return
	}

	var errMsg string
	if e.Err != nil {
		errMsg = e.Err.Error()
	} else {
		errMsg = "<nil>"
	}

	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s\n\tat %s:%d (%s) [package: %s]%s",
				errMsg, e.File, e.Line, e.Func, e.Package, e.StackTrace())
			return
		}
		fallthrough
	case 's':
		fmt.Fprintf(s, "%s\n\tat %s:%d (%s) [package: %s]",
			errMsg, e.File, e.Line, e.Func, e.Package)
	case 'q':
		fmt.Fprintf(s, "%q\n\tat %s:%d (%s) [package: %s]",
			errMsg, e.File, e.Line, e.Func, e.Package)
	}
}

// StackTrace returns the formatted stack trace if captured.
func (e *MetaError) StackTrace() string {
	if len(e.stackTrace) == 0 {
		return ""
	}
	var builder strings.Builder
	frames := runtime.CallersFrames(e.stackTrace)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "\n%s\n\t%s:%d", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	e.stackTraceString = builder.String()
	return e.stackTraceString
}

// Unwrap returns the underlying error.
func (e *MetaError) Unwrap() error {
	return e.Err
}

// Is delegates error comparison to the underlying error.
func (e *MetaError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// As delegates error type assertion to the underlying error.
func (e *MetaError) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

// RootCause returns the root cause of the error by unwrapping it until the end.
func RootCause(err error) error {
	for err != nil {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
	return nil
}

func (e *MetaError) ToCSV() string {
	record := []string{
		e.Err.Error(),
		e.File,
		strconv.Itoa(e.Line),
		e.Func,
		e.Package,
		//e.StackTrace(),
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	w.Comma = '|' // Use pipe as separator
	_ = w.Write(record)
	w.Flush()

	return strings.TrimSpace(buf.String())
}

func MetaErrorFromCSV(csvStr string) (*MetaError, error) {
	r := csv.NewReader(strings.NewReader(csvStr))
	r.Comma = '|' // Use pipe as separator
	r.FieldsPerRecord = 5

	record, err := r.Read()
	if err != nil {
		return nil, ErrNotMetaError
	}

	if len(record) != 5 {
		return nil, ErrNotMetaError
	}

	line, err := strconv.Atoi(record[2])
	if err != nil {
		return nil, ErrNotMetaError
	}

	return &MetaError{
		Err:     errors.New(record[0]),
		File:    record[1],
		Line:    line,
		Func:    record[3],
		Package: record[4],
		//stackTraceString: record[5],
	}, nil
}

func FromSlogMap(slogError map[string]interface{}) (*MetaError, error) {
	msgVal, ok := slogError["err"]
	if !ok {
		msgVal, ok = slogError["error"]
		if !ok {
			msgVal, ok = slogError["metaErr"]
			if !ok {
				return nil, ErrNotMetaError
			}
		}
	}

	metaErrorCSV, ok := msgVal.(string)
	if !ok {
		return nil, ErrNotMetaError
	}

	return MetaErrorFromCSV(metaErrorCSV)
}
