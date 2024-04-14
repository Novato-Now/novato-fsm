package journeystore

import (
	"context"
	"fmt"

	"github.com/Novato-Now/novato-fsm/errors"
	"github.com/Novato-Now/novato-fsm/model"
	novato_errors "github.com/Novato-Now/novato-utils/errors"
	"github.com/Novato-Now/novato-utils/logging"

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
	log := logging.GetLogger(ctx)

	jID := uuidNewString()
	log.Infof("Creating new journey with jID: %s", jID)

	journey := model.Journey[T]{
		JID: jID,
	}

	err := js.keyValueStore.Set(ctx, getJourneyKey(jID), journey)
	if err != nil {
		log.Errorf("Error creating new journey. Error: %+v", err)
		return model.Journey[T]{}, novato_errors.InternalSystemError(ctx)
	}
	log.Infof("Created new journey with jID: %s", jID)
	return journey, nil
}

func (js journeyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], *novato_errors.Error) {
	log := logging.GetLogger(ctx)

	log.Infof("Fetching journey with jID: %s", jID)
	journey, err := js.keyValueStore.Get(ctx, getJourneyKey(jID))

	if err != nil {
		log.Errorf("Error fetching journey. Error: %+v", err)
		return model.Journey[T]{}, novato_errors.InternalSystemError(ctx)
	}

	if journey == nil {
		log.Error("Journey does not exist.")
		return model.Journey[T]{}, errors.BypassError().WithMessage("journey not found")
	}

	return *journey, nil
}

func (js journeyStore[T]) Save(ctx context.Context, journey model.Journey[T]) *novato_errors.Error {
	log := logging.GetLogger(ctx)

	log.Infof("Saving journey with jID: %s", journey.JID)
	err := js.keyValueStore.Set(ctx, getJourneyKey(journey.JID), journey)
	if err != nil {
		log.Errorf("Error saving journey. Error: %+v", err)
		return novato_errors.InternalSystemError(ctx)
	}
	return nil
}

func (js journeyStore[T]) Delete(ctx context.Context, jID string) *novato_errors.Error {
	log := logging.GetLogger(ctx)

	log.Infof("Deleting journey with jID: %s", jID)
	err := js.keyValueStore.Del(ctx, getJourneyKey(jID))
	if err != nil {
		log.Errorf("Error deleting journey. Error: %+v", err)
		return novato_errors.InternalSystemError(ctx)
	}
	return nil
}

func getJourneyKey(jID string) string {
	return fmt.Sprintf("FSM_JOURNEY_%s", jID)
}
