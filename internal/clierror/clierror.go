package clierror

import (
	"errors"
	"net"

	"github.com/ferdikt/sensortower-cli/internal/exitcode"
	"github.com/ferdikt/sensortower-cli/internal/sensortower"
)

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

func Wrap(code int, message string) error {
	return &Error{Code: code, Message: message}
}

func Code(err error) int {
	if err == nil {
		return exitcode.OK
	}
	var cliErr *Error
	if errors.As(err, &cliErr) {
		return cliErr.Code
	}
	var httpErr *sensortower.HTTPError
	if errors.As(err, &httpErr) {
		switch {
		case httpErr.StatusCode == 429:
			return exitcode.RateLimit
		case httpErr.StatusCode == 401 || httpErr.StatusCode == 403:
			return exitcode.Auth
		case httpErr.StatusCode >= 400 && httpErr.StatusCode < 500:
			return exitcode.Validation
		default:
			return exitcode.Network
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return exitcode.Network
	}
	return exitcode.Internal
}
