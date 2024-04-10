package journeystore

import (
	"context"
	"fmt"

	fsmErrors "github.com/thevibegod/fsm/errors"
	"github.com/thevibegod/fsm/model"

	"github.com/google/uuid"
)

var uuidNewString = uuid.NewString

//go:generate mockgen -destination=../mocks/mock_journey_store.go -package=mocks -source=journey_store.go

type JourneyStore[T any] interface {
	Create(ctx context.Context) (model.Journey[T], *fsmErrors.FsmError)
	Get(ctx context.Context, jID string) (model.Journey[T], *fsmErrors.FsmError)
	Save(ctx context.Context, journey model.Journey[T]) *fsmErrors.FsmError
	Delete(ctx context.Context, jID string) *fsmErrors.FsmError
}

type journeyStore[T any] struct {
	keyValueStore KeyValueStore[T]
}

func NewJourneyStore[T any](keyValueStore KeyValueStore[T]) JourneyStore[T] {
	return journeyStore[T]{keyValueStore: keyValueStore}
}

func (js journeyStore[T]) Create(ctx context.Context) (model.Journey[T], *fsmErrors.FsmError) {
	jID := uuidNewString()
	journey := model.Journey[T]{
		JID: jID,
	}

	err := js.keyValueStore.Set(ctx, getJourneyKey(jID), journey)
	if err != nil {
		return model.Journey[T]{}, fsmErrors.InternalSystemError(err.Error())
	}
	return journey, nil
}

func (js journeyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], *fsmErrors.FsmError) {
	journey, err := js.keyValueStore.Get(ctx, getJourneyKey(jID))

	if err != nil {
		return model.Journey[T]{}, fsmErrors.InternalSystemError(err.Error())
	}

	if journey == nil {
		return model.Journey[T]{}, fsmErrors.ByPassError("journey not found")
	}

	return *journey, nil
}

func (js journeyStore[T]) Save(ctx context.Context, journey model.Journey[T]) *fsmErrors.FsmError {
	err := js.keyValueStore.Set(ctx, getJourneyKey(journey.JID), journey)
	if err != nil {
		return fsmErrors.InternalSystemError(err.Error())
	}
	return nil
}

func (js journeyStore[T]) Delete(ctx context.Context, jID string) *fsmErrors.FsmError {
	err := js.keyValueStore.Del(ctx, getJourneyKey(jID))
	if err != nil {
		return fsmErrors.InternalSystemError(err.Error())
	}
	return nil
}

func getJourneyKey(jID string) string {
	return fmt.Sprintf("FSM_JOURNEY_%s", jID)
}
