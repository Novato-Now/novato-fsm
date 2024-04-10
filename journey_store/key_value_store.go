package journeystore

import (
	"context"

	"github.com/thevibegod/fsm/model"
)

//go:generate mockgen -destination=../mocks/mock_key_value_store.go -package=mocks -source=key_value_store.go
type KeyValueStore[T any] interface {
	Set(ctx context.Context, key string, Value model.Journey[T]) error
	Get(ctx context.Context, key string) (*model.Journey[T], error)
	Del(ctx context.Context, key string) error
}
