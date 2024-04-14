package state_handler

import (
	"context"

	novato_errors "github.com/Novato-Now/novato-utils/errors"
)

//go:generate mockgen -destination=../mocks/mock_state_handler.go -package=mocks -source=state_handler.go

type StateHandler interface {
	Visit(ctx context.Context, jID string, journeyData any, data any) (response any, updatedJourneyData any, nextEvent string, err *novato_errors.Error)
	Revisit(ctx context.Context, jID string, journeyData any) (response any, updatedJourneyData any, err *novato_errors.Error)
}
