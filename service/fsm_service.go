package service

import (
	"context"
	"fmt"

	"github.com/thevibegod/fsm/constants"
	fsmErrors "github.com/thevibegod/fsm/errors"

	journeystore "github.com/thevibegod/fsm/journey_store"
	"github.com/thevibegod/fsm/model"
)

type FsmService[T interface{}] struct {
	states           map[string]model.FsmState
	initialStateName string
	finalStateName   string
	journeyStore     journeystore.JourneyStore[T]
}

func NewFsmService[T interface{}](initialState model.FsmState, nonInitStates []model.FsmState, journeyStore journeystore.JourneyStore[T]) (FsmService[T], *fsmErrors.FsmError) {
	fsmStateMap := make(map[string]model.FsmState)
	var finalStateName string
	for _, state := range nonInitStates {
		fsmStateMap[state.Name] = state
		if len(state.NextAvailableEvents) == 0 {
			if finalStateName != "" {
				return FsmService[T]{}, fsmErrors.InternalSystemError("multiple final states found")
			}
			finalStateName = state.Name
		}
	}

	if finalStateName == "" {
		return FsmService[T]{}, fsmErrors.InternalSystemError("no final state found")
	}

	fsmStateMap[initialState.Name] = initialState

	return FsmService[T]{states: fsmStateMap, journeyStore: journeyStore, initialStateName: initialState.Name, finalStateName: finalStateName}, nil
}

func (fs FsmService[T]) Execute(ctx context.Context, request model.FsmRequest) (model.FsmResponse, *fsmErrors.FsmError) {
	var journey model.Journey[T]
	var err *fsmErrors.FsmError
	if request.JID == "" {
		journey, err = fs.journeyStore.Create(ctx)
		if err != nil {
			return model.FsmResponse{}, err
		}
	} else {
		journey, err = fs.journeyStore.Get(ctx, request.JID)
		if err != nil {
			return model.FsmResponse{}, err
		}
	}

	for {
		currentStateName := journey.CurrentStage
		if currentStateName == "" {
			if request.Event != constants.EventNameStart {
				return model.FsmResponse{}, fsmErrors.ByPassError("invalid journey error: wrong event")
			}
			currentStateName = fs.initialStateName
			currentState, ok := fs.states[currentStateName]
			if !ok {
				return model.FsmResponse{}, fsmErrors.InternalSystemError("cannot find current state")
			}
			nextAvailableEvents := make(map[string]struct{})
			for _, nextPossibleEvent := range currentState.NextAvailableEvents {
				nextAvailableEvents[nextPossibleEvent.Event] = struct{}{}
			}
			response, updatedJourneyData, nextEvent, err := currentState.Action.Execute(ctx, journey.JID, journey.Data, request.Data, nextAvailableEvents)
			if err != nil {
				// Write rollback logic
				return model.FsmResponse{}, err
			}
			journey.Data = updatedJourneyData.(T)
			journey.CurrentStage = currentState.Name
			if currentState.IsCheckpoint {
				journey.LastCheckpointStage = currentState.Name
			}
			err = fs.journeyStore.Save(ctx, journey)
			if err != nil {
				// Write rollback logic
				return model.FsmResponse{}, err
			}
			if nextEvent == constants.EventNameTransitionComplete {
				return model.FsmResponse{JID: journey.JID, Data: response}, nil
			}
		}
		currentState, ok := fs.states[currentStateName]
		if !ok {
			return model.FsmResponse{}, fsmErrors.InternalSystemError("cannot find current state")
		}

		event := request.Event
		var nextState model.FsmState
		var nextStateFound bool
		for _, nextAvailableEvent := range currentState.NextAvailableEvents {
			if nextAvailableEvent.Event == event {
				nextState, nextStateFound = fs.states[nextAvailableEvent.DestinationStateName]
				if !nextStateFound {
					return model.FsmResponse{}, fsmErrors.InternalSystemError("cannot find next state")
				}
				nextAvailableEvents := make(map[string]struct{})
				for _, nextPossibleEvent := range nextState.NextAvailableEvents {
					nextAvailableEvents[nextPossibleEvent.Event] = struct{}{}
				}
				response, updatedJourneyData, nextEvent, err := nextState.Action.Execute(ctx, journey.JID, journey.Data, request.Data, nextAvailableEvents)
				if err != nil {
					// Write rollback logic
					return model.FsmResponse{}, err
				}
				journey.Data = updatedJourneyData.(T)
				journey.CurrentStage = nextState.Name
				if nextState.IsCheckpoint {
					journey.LastCheckpointStage = nextState.Name
				}
				err = fs.journeyStore.Save(ctx, journey)
				if err != nil {
					// Write rollback logic
					return model.FsmResponse{}, err
				}
				if nextEvent == constants.EventNameTransitionComplete {
					return model.FsmResponse{JID: journey.JID, Data: response}, nil
				}
			}
		}
		if !nextStateFound {
			return model.FsmResponse{}, fsmErrors.ByPassError(fmt.Sprintf("cannot find next state for event %s", request.Event))
		}
	}
}
