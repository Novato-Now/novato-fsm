package journeystore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/thevibegod/fsm/model"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type JourneyStore[T interface{}] struct {
	redisClient     redis.Client
	expiryInMinutes int
}

func (js JourneyStore[T]) Create(ctx context.Context, initStateName string) (model.Journey[T], error) {
	jID := uuid.NewString()
	journey := model.Journey[T]{
		JID:          jID,
		CurrentStage: initStateName,
	}

	bytes, _ := json.Marshal(journey)

	err := js.redisClient.Set(ctx, fmt.Sprintf("FSM_JOURNEY_%s", jID), string(bytes), time.Duration(js.expiryInMinutes*int(time.Minute))).Err()
	return journey, err
}

func (js JourneyStore[T]) Get(ctx context.Context, jID string) (model.Journey[T], error) {
	var journey model.Journey[T]
	journeyString, err := js.redisClient.Get(ctx, fmt.Sprintf("FSM_JOURNEY_%s", jID)).Result()

	if err != nil {
		return journey, err
	}

	err = json.Unmarshal([]byte(journeyString), &journey)

	return journey, err
}

func (js JourneyStore[T]) Save(ctx context.Context, journey model.Journey[T]) error {
	bytes, _ := json.Marshal(journey)

	err := js.redisClient.Set(ctx, fmt.Sprintf("FSM_JOURNEY_%s", journey.JID), string(bytes), time.Duration(js.expiryInMinutes*int(time.Minute))).Err()
	return err
}
