package logic

type ErrorCode string

const (
	ErrorCodeBadRequest   ErrorCode = "bad_request"
	ErrorCodeUnauthorized ErrorCode = "unauthorized"
	ErrorCodeForbidden    ErrorCode = "forbidden"
	ErrorCodeConflict     ErrorCode = "conflict"
	ErrorCodeNotFound     ErrorCode = "not_found"
	ErrorCodeUnavailable  ErrorCode = "service_unavailable"
)

type Error struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func badRequest(message string) Error {
	return Error{Code: ErrorCodeBadRequest, Message: message}
}

func unauthorized(message string) Error {
	return Error{Code: ErrorCodeUnauthorized, Message: message}
}

func forbidden(message string) Error {
	return Error{Code: ErrorCodeForbidden, Message: message}
}

func conflict(message string) Error {
	return Error{Code: ErrorCodeConflict, Message: message}
}

func notFound(message string) Error {
	return Error{Code: ErrorCodeNotFound, Message: message}
}

func serviceUnavailable(message string, cause error) Error {
	return Error{Code: ErrorCodeUnavailable, Message: message, Cause: cause}
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Cause
}
