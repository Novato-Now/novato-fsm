package action

import (
	"context"

	fsmErrors "github.com/thevibegod/fsm/errors"
)

type Action interface {
	Execute(ctx context.Context, jID string, journeyData interface{}, data interface{}) (response interface{}, updatedJourneyData interface{}, nextEvent string, err *fsmErrors.FsmError)
}
