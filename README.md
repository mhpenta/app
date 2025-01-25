# app

Go utilities for application development.

## Features

- `MetaError`: Rich error context with stack traces, file locations, and CSV serialization
- `MultiError`: Aggregate multiple errors
- `DebugContext`: Context with value inspection capabilities
- `CloseWithLog`: Resource cleanup with logging and retries
- Application mode control (`ReleaseMode`, `DevMode`, `DebugMode`)
- Context utilities
    - `MainContext`: App-level context with signal handling
    - `ContextCancelled`: Status checks
- Robustness helpers
    - `SleepMinPlusRandom`: Jittered delays
    - `ReturnTrueXPercentOfTime`: Probabilistic execution

## Usage

```go
// Error handling with context
err := app.NewMetaError(errors.New("something failed"))
fmt.Printf("%+v\n", err) // Prints error with stack trace

// Multiple errors
mErr := app.NewMultiError(err1, err2)
if mErr.HasErrors() {
    log.Fatal(mErr)
}

// Logging with context
defer app.LogSince("operation completed", time.Now())

// Resource cleanup
defer app.CloseWithLog(file, "file")
```

## Install

```bash
go get github.com/mhpenta/app
```