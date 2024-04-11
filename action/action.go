package action

import (
	"context"

	fsmErrors "github.com/thevibegod/fsm/errors"
)

//go:generate mockgen -destination=../mocks/mock_action.go -package=mocks -source=action.go

type Action interface {
	Execute(ctx context.Context, jID string, journeyData any, data any) (response any, updatedJourneyData any, nextEvent string, err *fsmErrors.FsmError)
}
