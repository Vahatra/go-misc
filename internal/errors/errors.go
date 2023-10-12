package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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

// Trusted error
type Error struct {
	err  error
	code Code
	// Time string `json:"timestamp"` // can be extended for richer error
}

func (e *Error) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func New(code Code, txt string) error {
	return &Error{code: code, err: errors.New(txt)}
}

func Newf(code Code, format string, a ...any) error {
	return &Error{err: errors.New(fmt.Sprintf(format, a...)), code: code}
}

func Wrap(code Code, err error) error {
	if err == nil {
		err = errors.New("")
	}
	return &Error{code: code, err: err}
}

func (e *Error) Code() Code {
	return e.code
}

// to satisfy grpc's {GRPCStatus() *Status}
func (e *Error) GRPCStatus() *status.Status {
	return status.New(GrpcCode(e.code), e.Error())
}

// Trusted errors
type Errors struct {
	errs []error
	code Code
}

func (e *Errors) Error() string {
	return strings.Join(e.Errors(), ", ")
}

func (e *Errors) Unwrap() []error {
	return e.errs
}

func (e *Errors) Errors() []string {
	s := make([]string, 0, len(e.errs))
	for _, err := range e.errs {
		s = append(s, err.Error())
	}
	return s
}

func NewS(code Code, s ...string) error {
	errs := make([]error, 0, len(s))
	for _, v := range s {
		errs = append(errs, errors.New(v))
	}
	return &Errors{code: code, errs: errs}
}

func WrapS(code Code, errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return &Errors{code: code}
	}
	e := &Errors{
		code: code,
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

func (e *Errors) Code() Code {
	return e.code
}

func (e *Errors) GRPCStatus() *status.Status {
	return status.New(GrpcCode(e.code), e.Error())
}

// helper function for getting the code out of an error,
// returns CodeUnknown if the err is not an Error/Errors
func GetCode(err error) Code {
	var ierr *Error
	if errors.As(err, &ierr) {
		return ierr.code
	}

	var ierrs *Errors
	if errors.As(err, &ierrs) {
		return ierrs.code
	}
	return CodeUnknown
}

func GrpcCode(code Code) codes.Code {
	switch code {
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

func HttpStatus(code Code) int {
	switch code {
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
