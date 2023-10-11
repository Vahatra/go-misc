package errors

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Code int

const (
	// http.StatusBadRequest
	// codes.InvalidArgument
	CodeInvalidArgument Code = iota
	// http.StatusNotFound
	// codes.NotFound
	CodeNotFound
	// http.StatusUnauthorized
	// codes.Unauthenticated
	CodeUnauthenticated
	// http.StatusForbidden
	// codes.PermissionDenied
	CodePermissionDenied
	// http.StatusInternalServerError
	// codes.Internal
	CodeInternal
	// http.StatusConflict
	// codes.AlreadyExists
	CodeAlreadyExists
	// http.StatusNotImplemented
	// codes.Unimplemented
	CodeUnimplemented
	// http.StatusInternalServerError
	// codes.Unknown
	CodeUnknown
)

var (
	ErrInternal error = &Error{code: CodeInternal}
	ErrNotFound error = &Error{code: CodeNotFound}
)

type Error struct {
	msg  string
	err  error
	code Code
	// Time string `json:"timestamp"` // can be extended for richer error
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Unwrap() error {
	return e.err
}

func New(code Code, msg string) error {
	return &Error{msg: msg, code: code}
}

func Newf(code Code, format string, a ...any) error {
	return &Error{msg: fmt.Sprintf(format, a...), code: code}
}

func Wrap(code Code, err error) error {
	var msg string
	if err != nil {
		msg = err.Error()
	}
	return &Error{err: err, msg: msg, code: code}
}

func (e *Error) Code() Code {
	return e.code
}

// helper function for getting the code out of an error
// returns CodeUnknown if the err is not an Error
func GetCode(err error) Code {
	var ierr *Error
	if errors.As(err, &ierr) {
		return ierr.code
	}
	return CodeUnknown
}

// to satisfy grpc's {GRPCStatus() *Status}
func (e *Error) GRPCStatus() *status.Status {
	return status.New(e.ToGRPCCode(), e.msg)
}

func (e *Error) ToGRPCCode() codes.Code {
	switch e.code {
	case CodeInvalidArgument:
		return codes.InvalidArgument
	case CodeNotFound:
		return codes.NotFound
	case CodeUnauthenticated:
		return codes.Unauthenticated
	case CodePermissionDenied:
		return codes.PermissionDenied
	case CodeInternal:
		return codes.Internal
	case CodeAlreadyExists:
		return codes.AlreadyExists
	case CodeUnimplemented:
		return codes.Unimplemented
	default:
		return codes.Unknown
	}
}

func (e *Error) ToHTTPStatus() int {
	switch e.code {
	case CodeInvalidArgument:
		return http.StatusBadRequest
	case CodeNotFound:
		return http.StatusNotFound
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodePermissionDenied:
		return http.StatusForbidden
	case CodeInternal:
		return http.StatusInternalServerError
	case CodeAlreadyExists:
		return http.StatusConflict
	case CodeUnimplemented:
		return http.StatusNotImplemented
	case CodeUnknown:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
