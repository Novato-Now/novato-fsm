package journeystore

import (
	"context"
	"fmt"

	novato_errors "github.com/Novato-Now/novato-utils/errors"
	"github.com/thevibegod/fsm/errors"
	"github.com/thevibegod/fsm/model"

	"github.com/google/uuid"
)

var uuidNewString = uuid.NewString

//go:generate mockgen -destination=../mocks/mock_journey_store.go -package=mocks -source=journey_store.go

type JourneyStore[T any] interface {
	Create(ctx context.Context) (model.Journey[T], *novato_errors.Error)
	Get(ctx context.Context, jID string) (model.Journey[T], *novato_errors.Error)
	Save(ctx context.Context, journey model.Journey[T]) *novato_errors.Error
	Delete(ctx context.Context, jID string) *novato_errors.Error
}

type journeyStore[T any] struct {
	keyValueStore KeyValueStore[T]
}

func NewJourneyStore[T any](keyValueStore KeyValueStore[T]) JourneyStore[T] {
	return journeyStore[T]{keyValueStore: keyValueStore}
}

func (js journeyStore[T]) Create(ctx context.Context) (model.Journey[T], *novato_errors.Error) {
	jID := uuidNewString()
	journey := model.Journey[T]{
		JID: jID,
	}

	err := js.keyValueStore.Set(ctx, getJourneyKey(jID), journey)
	if err != nil {
		return model.Journey[T]{}, novato_errors.InternalSystemError(ctx).WithMessage(err.Error())
	}
	return journey, nil
}

func (js journeyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], *novato_errors.Error) {
	journey, err := js.keyValueStore.Get(ctx, getJourneyKey(jID))

	if err != nil {
		return model.Journey[T]{}, novato_errors.InternalSystemError(ctx).WithMessage(err.Error())
	}

	if journey == nil {
		return model.Journey[T]{}, errors.BypassError().WithMessage("journey not found")
	}

	return *journey, nil
}

func (js journeyStore[T]) Save(ctx context.Context, journey model.Journey[T]) *novato_errors.Error {
	err := js.keyValueStore.Set(ctx, getJourneyKey(journey.JID), journey)
	if err != nil {
		return novato_errors.InternalSystemError(ctx).WithMessage(err.Error())
	}
	return nil
}

func (js journeyStore[T]) Delete(ctx context.Context, jID string) *novato_errors.Error {
	err := js.keyValueStore.Del(ctx, getJourneyKey(jID))
	if err != nil {
		return novato_errors.InternalSystemError(ctx).WithMessage(err.Error())
	}
	return nil
}

func getJourneyKey(jID string) string {
	return fmt.Sprintf("FSM_JOURNEY_%s", jID)
}
