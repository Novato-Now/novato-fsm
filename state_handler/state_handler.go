package state_handler

import (
	"context"

	fsmErrors "github.com/thevibegod/fsm/errors"
)

//go:generate mockgen -destination=../mocks/mock_state_handler.go -package=mocks -source=state_handler.go

type StateHandler interface {
	Visit(ctx context.Context, jID string, journeyData any, data any) (response any, updatedJourneyData any, nextEvent string, err *fsmErrors.FsmError)
	Revisit(ctx context.Context, jID string, journeyData any) (response any, updatedJourneyData any, err *fsmErrors.FsmError)
}
