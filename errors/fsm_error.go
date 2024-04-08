package errors

import (
	"fmt"
	"net/http"
)

type FsmError struct {
	ErrorCode  string
	ErrorMsg   string
	StatusCode int
}

func (err *FsmError) Error() string {
	return fmt.Sprintf("{ErrorCode: \"%s\", ErrorMsg: \"%s\"}", err.ErrorCode, err.ErrorMsg)
}

func ByPassError(errorMsg string) *FsmError {
	return &FsmError{
		ErrorCode:  "FSM_BYPASS_ERROR",
		ErrorMsg:   errorMsg,
		StatusCode: http.StatusForbidden,
	}
}

func InternalSystemError(errorMsg string) *FsmError {
	return &FsmError{
		ErrorCode: "FSM_INTERNAL_SYSTEM_ERROR",
		ErrorMsg:  errorMsg,
		StatusCode: http.StatusInternalServerError,
	}
}

func BadRequestError(errorMsg string) *FsmError {
	return &FsmError{
		ErrorCode: "FSM_BAD_REQUEST_ERROR",
		ErrorMsg:  errorMsg,
		StatusCode: http.StatusBadRequest,
	}
}

func DependencySystemError(errorMsg string) *FsmError {
	return &FsmError{
		ErrorCode: "FSM_DEPENDENCY_SYSTEM_ERROR",
		ErrorMsg:  errorMsg,
		StatusCode: http.StatusFailedDependency,
	}
}
