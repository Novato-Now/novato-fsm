package service

import (
	"context"
	"fmt"

	"github.com/thevibegod/fsm/constants"
	fsmErrors "github.com/thevibegod/fsm/errors"

	journeystore "github.com/thevibegod/fsm/journey_store"
	"github.com/thevibegod/fsm/model"
)

//go:generate mockgen -destination=../mocks/mock_fsm_service.go -package=mocks -source=fsm_service.go

type FsmService[T any] interface {
	Execute(ctx context.Context, request model.FsmRequest) (response model.FsmResponse, err *fsmErrors.FsmError)
}

type fsmService[T any] struct {
	states           map[string]model.FsmState
	initialStateName string
	finalStateName   string
	journeyStore     journeystore.JourneyStore[T]
}

func NewFsmService[T any](initialState model.FsmState, nonInitStates []model.FsmState, journeyStore journeystore.JourneyStore[T]) (FsmService[T], *fsmErrors.FsmError) {
	fsmStateMap := make(map[string]model.FsmState)
	var finalStateName string
	for _, state := range nonInitStates {
		fsmStateMap[state.Name] = state
		if len(state.NextAvailableEvents) == 0 {
			if finalStateName != "" {
				return fsmService[T]{}, fsmErrors.InternalSystemError("multiple final states found")
			}
			finalStateName = state.Name
		}
	}

	if finalStateName == "" {
		return fsmService[T]{}, fsmErrors.InternalSystemError("no final state found")
	}

	fsmStateMap[initialState.Name] = initialState

	return fsmService[T]{states: fsmStateMap, journeyStore: journeyStore, initialStateName: initialState.Name, finalStateName: finalStateName}, nil
}

func (fs fsmService[T]) Execute(ctx context.Context, request model.FsmRequest) (response model.FsmResponse, err *fsmErrors.FsmError) {
	var journey model.Journey[T]

	var currentState, nextState, lastExecutedState model.FsmState
	var nextStateData any
	var nextEvent string

	var isNewJourney bool
	var finishStateTransition bool

	defer func() {
		if err != nil && isNewJourney {
			_ = fs.journeyStore.Delete(ctx, journey.JID)
		}
	}()

	if request.JID != "" {
		journey, err = fs.journeyStore.Get(ctx, request.JID)
		if err != nil {
			return
		}
		nextStateData = request.Data
		nextEvent = request.Event
	} else {
		isNewJourney = true
		journey, nextStateData, nextEvent, err = fs.startNewJourney(ctx, request.Data, request.Event)
		if err != nil {
			return
		}
		lastExecutedState, err = fs.getState(journey.CurrentStage)
		if err != nil {
			return
		}
		if nextEvent == constants.EventNameTransitionComplete {
			finishStateTransition = true
		}
	}

	for !finishStateTransition {
		currentState, err = fs.getState(journey.CurrentStage)
		if err != nil {
			return
		}
		nextState, err = fs.getNextState(currentState, nextEvent)
		if err != nil {
			return
		}
		journey, nextStateData, nextEvent, err = fs.executeAction(ctx, nextState, journey, nextStateData)
		if err != nil {
			return
		}
		lastExecutedState = nextState
		if nextEvent == constants.EventNameTransitionComplete {
			finishStateTransition = true
		}
	}

	err = fs.journeyStore.Save(ctx, journey)
	if err != nil {
		return
	}

	return fs.loadFsmResponse(journey, lastExecutedState, nextStateData), nil
}

func (fs fsmService[T]) getState(stateName string) (model.FsmState, *fsmErrors.FsmError) {
	state, ok := fs.states[stateName]
	if !ok {
		return model.FsmState{}, fsmErrors.InternalSystemError("cannot find next state")
	}
	return state, nil
}

func (fs fsmService[T]) getNextState(currentState model.FsmState, event string) (model.FsmState, *fsmErrors.FsmError) {
	for _, nextAvailableEvent := range currentState.NextAvailableEvents {
		if nextAvailableEvent.Event == event {
			return fs.getState(nextAvailableEvent.DestinationStateName)
		}
	}
	return model.FsmState{}, fsmErrors.ByPassError(fmt.Sprintf("invalid event %s for state %s", event, currentState.Name))
}

func (fs fsmService[T]) executeAction(ctx context.Context, state model.FsmState, journey model.Journey[T], data any) (model.Journey[T], any, string, *fsmErrors.FsmError) {
	resp, updatedJourneyData, nextEvent, err := state.Action.Execute(ctx, journey.JID, journey.Data, data)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	journey.Data = updatedJourneyData.(T)
	journey.CurrentStage = state.Name
	if state.IsCheckpoint {
		journey.LastCheckpointStage = state.Name
	}
	return journey, resp, nextEvent, nil
}

func (fs fsmService[T]) startNewJourney(ctx context.Context, data any, event string) (model.Journey[T], any, string, *fsmErrors.FsmError) {
	if event != constants.EventNameStart {
		return model.Journey[T]{}, nil, "", fsmErrors.ByPassError("invalid journey error: wrong event")
	}
	initState, err := fs.getState(fs.initialStateName)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	journey, err := fs.journeyStore.Create(ctx)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	journey, resp, nextEvent, err := fs.executeAction(ctx, initState, journey, data)
	if err != nil {
		_ = fs.journeyStore.Delete(ctx, journey.JID)
		return model.Journey[T]{}, nil, "", err
	}

	return journey, resp, nextEvent, nil
}

func (fs fsmService[T]) loadFsmResponse(journey model.Journey[T], state model.FsmState, response any) model.FsmResponse {
	return model.FsmResponse{
		JID:        journey.JID,
		Data:       response,
		NextScreen: state.NextScreen,
		MetaData:   state.MetaData,
	}
}
