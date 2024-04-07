package action

import "context"

type Action interface {
	Execute(ctx context.Context, jID string, journeyData interface{}, data interface{}, nextAvailableEvents map[string]struct{}) (response interface{}, updatedJourneyData interface{}, nextEvent string, err error)
}
