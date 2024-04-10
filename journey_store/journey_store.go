package journeystore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	fsmErrors "github.com/thevibegod/fsm/errors"
	"github.com/thevibegod/fsm/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JourneyStore[T interface{}] struct {
	redisClient     redis.Client
	expiryInMinutes int
}

func NewJourneyStore[T interface{}](redisClient redis.Client, expiryInMinutes int) JourneyStore[T] {
	return JourneyStore[T]{redisClient: redisClient, expiryInMinutes: expiryInMinutes}
}

func (js JourneyStore[T]) Create(ctx context.Context) (model.Journey[T], *fsmErrors.FsmError) {
	jID := uuid.NewString()
	journey := model.Journey[T]{
		JID: jID,
	}

	bytes, _ := json.Marshal(journey)

	err := js.redisClient.Set(ctx, getJourneyKey(jID), string(bytes), time.Duration(js.expiryInMinutes*int(time.Minute))).Err()
	if err != nil {
		return journey, fsmErrors.InternalSystemError(err.Error())
	}
	return journey, nil
}

func (js JourneyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], *fsmErrors.FsmError) {
	var journey model.Journey[T]
	journeyString, err := js.redisClient.Get(ctx, getJourneyKey(jID)).Result()

	if errors.Is(err, redis.Nil) {
		return journey, fsmErrors.ByPassError("journey not found")
	}

	if err != nil {
		return journey, fsmErrors.InternalSystemError(err.Error())
	}

	err = json.Unmarshal([]byte(journeyString), &journey)
	if err != nil {
		return journey, fsmErrors.InternalSystemError(err.Error())
	}

	return journey, nil
}

func (js JourneyStore[T]) Save(ctx context.Context, journey model.Journey[T]) *fsmErrors.FsmError {
	bytes, _ := json.Marshal(journey)

	err := js.redisClient.Set(ctx, getJourneyKey(journey.JID), string(bytes), time.Duration(js.expiryInMinutes*int(time.Minute))).Err()
	if err != nil {
		return fsmErrors.InternalSystemError(err.Error())
	}
	return nil
}

func (js JourneyStore[T]) Delete(ctx context.Context, jID string) *fsmErrors.FsmError {

	err := js.redisClient.Del(ctx, getJourneyKey(jID)).Err()
	if err != nil {
		return fsmErrors.InternalSystemError(err.Error())
	}
	return nil
}

func getJourneyKey(jID string) string {
	return fmt.Sprintf("FSM_JOURNEY_%s", jID)
}
