package apierr

import "fmt"

type WorkbenchAPIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *WorkbenchAPIError) Error() string {
	return e.Message
}

func InvalidArgument(message string) error {
	return &WorkbenchAPIError{
		Code:       "INVALID_ARGUMENT",
		Message:    message,
		HTTPStatus: 400,
	}
}

func Unauthorized(message string) error {
	return &WorkbenchAPIError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		HTTPStatus: 401,
	}
}

func NotFound(message string) error {
	return &WorkbenchAPIError{
		Code:       "NOT_FOUND",
		Message:    message,
		HTTPStatus: 404,
	}
}

func State(message string) error {
	return &WorkbenchAPIError{
		Code:       "STATE_ERROR",
		Message:    message,
		HTTPStatus: 503,
	}
}

func WrapInternal(message string, err error) error {
	if err == nil {
		return &WorkbenchAPIError{
			Code:       "INTERNAL_ERROR",
			Message:    message,
			HTTPStatus: 500,
		}
	}
	return &WorkbenchAPIError{
		Code:       "INTERNAL_ERROR",
		Message:    fmt.Sprintf("%s: %v", message, err),
		HTTPStatus: 500,
	}
}

func AsAPIError(err error) *WorkbenchAPIError {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*WorkbenchAPIError); ok {
		return apiErr
	}
	return &WorkbenchAPIError{
		Code:       "INTERNAL_ERROR",
		Message:    err.Error(),
		HTTPStatus: 500,
	}
}
