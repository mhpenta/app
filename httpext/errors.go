package httpext

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"strings"
	"syscall"
)

const (
	InternalServerError  = "internal server error"
	BadRequestError      = "bad request error"
	possibleConnResetMsg = "connection reset by peer"
	possibleGotAwayMsg   = "server sent GOAWAY"
)

// IsTransientNetworkOrDNSIssueErr checks if the error is a possible network or DNS issue.
func IsTransientNetworkOrDNSIssueErr(err error) bool {
	if err == nil {
		return false
	}

	// Unwrap the error to get the root cause
	unwrappedErr := errors.Unwrap(err)
	if unwrappedErr != nil {
		err = unwrappedErr
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "connection reset by peer") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "connection timed out") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "tls handshake timeout") ||
		strings.Contains(errMsg, "temporary failure in name resolution") ||
		strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "connection closed") ||
		strings.Contains(errMsg, "http2: server sent goaway") ||
		strings.Contains(errMsg, "unexpected eof") ||
		strings.Contains(errMsg, "server misbehaving") {
		return true
	}
	return false
}

// IsDialError determines if the given error is related to network dialing or connectivity issues.
// It checks for various types of network errors, including:
//   - Timeout errors (net.Error with Timeout() == true)
//   - Dial and read operation errors (net.OpError)
//   - Specific system errors like connection refused, host unreachable, and network unreachable
//   - DNS lookup timeout errors (net.DNSError)
//   - Generic timeout errors (detected by os.IsTimeout)
//   - String matching for common network error messages
//
// This function is useful for determining if an error is likely due to network issues
// and may be resolved by retrying the operation after a delay.
//
// Returns true if the error is identified as a network dialing or connectivity issue,
// false otherwise or if the input error is nil.
func IsDialError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" || opErr.Op == "read" {
			return true
		}

		var sysErr syscall.Errno
		if errors.As(opErr.Err, &sysErr) {
			switch sysErr {
			case syscall.ECONNREFUSED, syscall.EHOSTUNREACH, syscall.ENETUNREACH, syscall.ETIMEDOUT:
				return true
			}
		}
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		slog.Warn("DNS lookup error encountered",
			"error", dnsErr,
			"isTimeout", dnsErr.IsTimeout,
			"name", dnsErr.Name)
		return dnsErr.IsTimeout
	}

	if os.IsTimeout(err) {
		return true
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "network is unreachable") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "i/o timeout")
}

// IsConnectionResetByPeerError determines if the given error is a connection reset by peer error.
func IsConnectionResetByPeerError(err error) bool {
	// You'd think this would be formally defined somewhere but search the string in
	// the std library and you'll find it is a string literal in the sys package
	if strings.Contains(err.Error(), possibleConnResetMsg) {
		return true
	}
	return false
}

// IsHTTP2GoAwayError determines if the given error is a http2 GOAWAY error.
//
//	The GOAWAY frame (type=0x7) is used to initiate shutdown of a
//	connection or to signal serious error conditions.  GOAWAY allows an
//	endpoint to gracefully stop accepting new streams while still
//	finishing processing of previously established streams.  This enables
//	administrative actions, like server maintenance." From RFC7540
//
// See https://datatracker.ietf.org/doc/html/rfc7540#section-6.8
func IsHTTP2GoAwayError(err error) bool {
	if strings.Contains(err.Error(), possibleGotAwayMsg) {
		return true
	}
	return false
}

// IsIOTimeoutError determines if the given error is an I/O timeout error.
// It checks for various types of timeout errors, including:
//   - net.Error with Timeout() == true
//   - net.OpError with Op == "read" or "write" and containing timeout
//   - Generic os.IsTimeout errors
//   - String matching for common I/O timeout error messages
//
// Returns true if the error is identified as an I/O timeout error,
// false otherwise or if the input error is nil.
func IsIOTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if (opErr.Op == "read" || opErr.Op == "write") &&
			(opErr.Timeout() || strings.Contains(opErr.Error(), "i/o timeout")) {
			return true
		}
	}

	if os.IsTimeout(err) {
		return true
	}

	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "i/o timeout") ||
		strings.Contains(errMsg, "operation timed out") ||
		strings.Contains(errMsg, "read timeout") ||
		strings.Contains(errMsg, "write timeout")
}
