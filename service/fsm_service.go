package service

import (
	"context"
	"errors"
	"fmt"
	journeystore "fsm/journey_store"
	"fsm/model"
)

type FsmService[T interface{}] struct {
	states           map[string]model.FsmState
	initialStateName string
	finalStateName   string
	journeyStore     journeystore.JourneyStore[T]
}

func NewFsmService[T interface{}](initialState model.FsmState, nonInitStates []model.FsmState, journeyStore journeystore.JourneyStore[T]) (FsmService[T], error) {
	fsmStateMap := make(map[string]model.FsmState)
	var finalStateName string
	for _, state := range nonInitStates {
		fsmStateMap[state.Name] = state
		if len(state.NextAvailableEvents) == 0 {
			if finalStateName != "" {
				return FsmService[T]{}, errors.New("multiple final states found")
			}
			finalStateName = state.Name
		}
	}

	if finalStateName == "" {
		return FsmService[T]{}, errors.New("no final state found")
	}

	fsmStateMap[initialState.Name] = initialState

	return FsmService[T]{states: fsmStateMap, journeyStore: journeyStore, initialStateName: initialState.Name, finalStateName: finalStateName}, nil
}

func (fs FsmService[T]) Execute(ctx context.Context, request model.FsmRequest) (model.FsmResponse, error) {
	var journey model.Journey[T]
	var err error
	if request.JID == "" {
		journey, err = fs.journeyStore.Create(ctx, fs.initialStateName)
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
		currentState, ok := fs.states[currentStateName]
		if !ok {
			return model.FsmResponse{}, errors.New("invalid journey error: cannot find current state")
		}

		event := request.Event
		var nextState model.FsmState
		var nextStateFound bool
		for _, nextAvailableEvent := range currentState.NextAvailableEvents {
			if nextAvailableEvent.Event == event {
				nextState, nextStateFound = fs.states[nextAvailableEvent.DestinationStateName]
				if !nextStateFound {
					return model.FsmResponse{}, errors.New("invalid journey error: cannot find next state")
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
				if nextEvent == "TransitionComplete" {
					return model.FsmResponse{JID: journey.JID, Data: response}, nil
				}
			}
		}
		if !nextStateFound {
			return model.FsmResponse{}, fmt.Errorf("invalid journey error: cannot find next state for event %s", request.Event)
		}
	}
}
