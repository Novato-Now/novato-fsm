package service

import (
	"context"
	"fmt"

	"github.com/Novato-Now/novato-fsm/constants"
	fsmErrors "github.com/Novato-Now/novato-fsm/errors"
	nuConstants "github.com/Novato-Now/novato-utils/constants"
	nuErrors "github.com/Novato-Now/novato-utils/errors"

	journeystore "github.com/Novato-Now/novato-fsm/journey_store"
	"github.com/Novato-Now/novato-fsm/model"
)

//go:generate mockgen -destination=../mocks/mock_fsm_service.go -package=mocks -source=fsm_service.go

type FsmService[T any] interface {
	Execute(ctx context.Context, request model.FsmRequest) (response model.FsmResponse, err *nuErrors.Error)
}

type fsmService[T any] struct {
	states           map[string]model.FsmState
	initialStateName string
	finalStateName   string
	journeyStore     journeystore.JourneyStore[T]
}

func NewFsmService[T any](initialState model.FsmState, nonInitStates []model.FsmState, journeyStore journeystore.JourneyStore[T]) (FsmService[T], *nuErrors.Error) {
	ctx := context.WithValue(context.Background(), nuConstants.ServiceNameKey, "FSM")
	fsmStateMap := make(map[string]model.FsmState)
	var finalStateName string
	for _, state := range nonInitStates {
		fsmStateMap[state.Name] = state
		if len(state.NextAvailableEvents) == 0 {
			if finalStateName != "" {
				return fsmService[T]{}, nuErrors.InternalSystemError(ctx).WithMessage("multiple final states found")
			}
			finalStateName = state.Name
		}
	}

	if finalStateName == "" {
		return fsmService[T]{}, nuErrors.InternalSystemError(ctx).WithMessage("no final state found")
	}

	fsmStateMap[initialState.Name] = initialState

	return fsmService[T]{states: fsmStateMap, journeyStore: journeyStore, initialStateName: initialState.Name, finalStateName: finalStateName}, nil
}

func (fs fsmService[T]) Execute(ctx context.Context, request model.FsmRequest) (response model.FsmResponse, err *nuErrors.Error) {
	var journey model.Journey[T]

	var currentState, nextState, lastExecutedState model.FsmState
	var nextStateData any
	var nextEvent string

	var finishStateTransition bool

	if request.JID != "" {
		journey, err = fs.journeyStore.Get(ctx, request.JID)
		if err != nil {
			return
		}
		if request.Event == constants.EventNameResume {
			response, err = fs.handleResumeJourney(ctx, journey)
			return
		}
		if request.Event == constants.EventNameBack {
			response, err = fs.handleBackJourney(ctx, journey)
			return
		}
		nextStateData = request.Data
		nextEvent = request.Event
	} else {
		journey, nextStateData, nextEvent, err = fs.startNewJourney(ctx, request.Data, request.Event)
		if err != nil {
			return
		}
		defer func() {
			if err != nil {
				_ = fs.journeyStore.Delete(ctx, journey.JID)
			}
		}()
		lastExecutedState, err = fs.getState(ctx, journey.CurrentStage)
		if err != nil {
			return
		}
		if nextEvent == constants.EventNameTransitionComplete {
			finishStateTransition = true
		}
	}

	for !finishStateTransition {
		currentState, err = fs.getState(ctx, journey.CurrentStage)
		if err != nil {
			return
		}
		nextState, err = fs.getNextState(ctx, currentState, nextEvent)
		if err != nil {
			return
		}
		journey, nextStateData, nextEvent, err = fs.handleStateVisit(ctx, nextState, journey, nextStateData)
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

func (fs fsmService[T]) getState(ctx context.Context, stateName string) (model.FsmState, *nuErrors.Error) {
	state, ok := fs.states[stateName]
	if !ok {
		return model.FsmState{}, nuErrors.InternalSystemError(ctx).WithMessage("cannot find next state")
	}
	return state, nil
}

func (fs fsmService[T]) getNextState(ctx context.Context, currentState model.FsmState, event string) (model.FsmState, *nuErrors.Error) {
	for _, nextAvailableEvent := range currentState.NextAvailableEvents {
		if nextAvailableEvent.Event == event {
			return fs.getState(ctx, nextAvailableEvent.DestinationStateName)
		}
	}
	return model.FsmState{}, fsmErrors.BypassError().WithMessage(fmt.Sprintf("invalid event %s for state %s", event, currentState.Name))
}

func (fs fsmService[T]) handleStateVisit(ctx context.Context, state model.FsmState, journey model.Journey[T], data any) (model.Journey[T], any, string, *nuErrors.Error) {
	resp, updatedJourneyData, nextEvent, err := state.StateHandler.Visit(ctx, journey.JID, journey.Data, data)
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

func (fs fsmService[T]) handleStateRevisit(ctx context.Context, state model.FsmState, journey model.Journey[T]) (model.Journey[T], any, *nuErrors.Error) {
	resp, updatedJourneyData, err := state.StateHandler.Revisit(ctx, journey.JID, journey.Data)
	if err != nil {
		return model.Journey[T]{}, nil, err
	}
	journey.CurrentStage = state.Name
	if state.IsCheckpoint {
		journey.LastCheckpointStage = state.Name
	}
	journey.Data = updatedJourneyData.(T)
	return journey, resp, nil
}

func (fs fsmService[T]) handleResumeJourney(ctx context.Context, journey model.Journey[T]) (model.FsmResponse, *nuErrors.Error) {
	state, err := fs.getState(ctx, journey.LastCheckpointStage)
	if err != nil {
		return model.FsmResponse{}, err
	}
	return fs.revisitAndSave(ctx, journey, state)
}

func (fs fsmService[T]) handleBackJourney(ctx context.Context, journey model.Journey[T]) (model.FsmResponse, *nuErrors.Error) {
	state, err := fs.getState(ctx, journey.CurrentStage)
	if err != nil {
		return model.FsmResponse{}, err
	}
	nextState, err := fs.getNextState(ctx, state, constants.EventNameBack)
	if err != nil {
		return model.FsmResponse{}, err
	}
	return fs.revisitAndSave(ctx, journey, nextState)
}

func (fs fsmService[T]) startNewJourney(ctx context.Context, data any, event string) (model.Journey[T], any, string, *nuErrors.Error) {
	if event != constants.EventNameStart {
		return model.Journey[T]{}, nil, "", fsmErrors.BypassError().WithMessage("invalid journey error: wrong event")
	}
	initState, err := fs.getState(ctx, fs.initialStateName)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	journey, err := fs.journeyStore.Create(ctx)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	jid := journey.JID
	journey, resp, nextEvent, err := fs.handleStateVisit(ctx, initState, journey, data)
	if err != nil {
		_ = fs.journeyStore.Delete(ctx, jid)
		return model.Journey[T]{}, nil, "", err
	}

	return journey, resp, nextEvent, nil
}

func (fs fsmService[T]) revisitAndSave(ctx context.Context, journey model.Journey[T], state model.FsmState) (model.FsmResponse, *nuErrors.Error) {
	journey, resp, err := fs.handleStateRevisit(ctx, state, journey)
	if err != nil {
		return model.FsmResponse{}, err
	}
	err = fs.journeyStore.Save(ctx, journey)
	if err != nil {
		return model.FsmResponse{}, err
	}
	return fs.loadFsmResponse(journey, state, resp), nil
}

func (fs fsmService[T]) loadFsmResponse(journey model.Journey[T], state model.FsmState, response any) model.FsmResponse {
	return model.FsmResponse{
		JID:        journey.JID,
		Data:       response,
		NextScreen: state.NextScreen,
		MetaData:   state.MetaData,
	}
}
