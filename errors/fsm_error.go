package errors

import (
	"net/http"

	novato_errors "github.com/Novato-Now/novato-utils/errors"
)

func BypassError() *novato_errors.Error {
	return novato_errors.New("FSM_BYPASS_ERROR", http.StatusForbidden)
}
