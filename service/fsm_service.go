package service

import (
	"context"

	"github.com/Novato-Now/novato-fsm/constants"
	fsmErrors "github.com/Novato-Now/novato-fsm/errors"
	nuConstants "github.com/Novato-Now/novato-utils/constants"
	nuErrors "github.com/Novato-Now/novato-utils/errors"
	"github.com/Novato-Now/novato-utils/logging"

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
	log := logging.GetLogger(ctx)
	var journey model.Journey[T]

	var currentState, nextState, lastExecutedState model.FsmState
	var nextStateData any
	var nextEvent string

	var finishStateTransition bool

	if request.JID != "" {
		log.Info("Journey id found. Fetching journey from journey store.")
		journey, err = fs.journeyStore.Get(ctx, request.JID)
		if err != nil {
			log.Errorf("Error from journey store. Error %+v", err)
			return
		}
		if request.Event == constants.EventNameResume {
			log.Info("Found resume event.")
			response, err = fs.handleResumeJourney(ctx, journey)
			return
		}
		if request.Event == constants.EventNameBack {
			log.Info("Found back event.")
			response, err = fs.handleBackJourney(ctx, journey)
			return
		}
		nextStateData = request.Data
		nextEvent = request.Event
	} else {
		log.Info("No journey id found.")
		if request.Event != constants.EventNameStart {
			log.Error("Invalid event name for new journey.")
			err = fsmErrors.BypassError()
			return
		}
		log.Info("Journey id not found. Starting new journey.")
		journey, nextStateData, nextEvent, err = fs.startNewJourney(ctx, request.Data)
		if err != nil {
			log.Errorf("Unable to start new journey. Error: %+v", err)
			return
		}
		defer func() {
			if err != nil {
				log.Info("Rolling back journey creation")
				deleteErr := fs.journeyStore.Delete(ctx, journey.JID)
				if deleteErr != nil {
					log.Warnf("Unable to delete journey for JID %s", journey.JID)
				}
			}
		}()
		lastExecutedState, err = fs.getState(ctx, journey.CurrentStage)
		if err != nil {
			log.Errorf("Unable to fetch last executed state. Error: %+v", err)
			return
		}
		if nextEvent == constants.EventNameTransitionComplete {
			finishStateTransition = true
		}
	}

	for !finishStateTransition {
		currentState, err = fs.getState(ctx, journey.CurrentStage)
		if err != nil {
			log.Errorf("Unable to fetch current state. Error: %+v", err)
			return
		}
		nextState, err = fs.getNextState(ctx, currentState, nextEvent)
		if err != nil {
			log.Errorf("Unable to fetch next state. Error: %+v", err)
			return
		}
		journey, nextStateData, nextEvent, err = fs.handleStateVisit(ctx, nextState, journey, nextStateData)
		if err != nil {
			log.Errorf("Error from state handler visit. Error: %+v", err)
			return
		}
		lastExecutedState = nextState
		if nextEvent == constants.EventNameTransitionComplete {
			finishStateTransition = true
		}
	}

	err = fs.journeyStore.Save(ctx, journey)
	if err != nil {
		log.Errorf("Unable to save journey. Error: %+v", err)
		return
	}

	return fs.loadFsmResponse(journey, lastExecutedState, nextStateData), nil
}

func (fs fsmService[T]) getState(ctx context.Context, stateName string) (model.FsmState, *nuErrors.Error) {
	log := logging.GetLogger(ctx)
	state, ok := fs.states[stateName]
	if !ok {
		log.Errorf("Cannot find state with name %s", stateName)
		return model.FsmState{}, nuErrors.InternalSystemError(ctx)
	}
	return state, nil
}

func (fs fsmService[T]) getNextState(ctx context.Context, currentState model.FsmState, event string) (model.FsmState, *nuErrors.Error) {
	log := logging.GetLogger(ctx)
	for _, nextAvailableEvent := range currentState.NextAvailableEvents {
		if nextAvailableEvent.Event == event {
			log.Infof("Found next state as %s", nextAvailableEvent.DestinationStateName)
			return fs.getState(ctx, nextAvailableEvent.DestinationStateName)
		}
	}
	log.Errorf("Invalid event %s for state %s", event, currentState.Name)
	return model.FsmState{}, fsmErrors.BypassError()
}

func (fs fsmService[T]) handleStateVisit(ctx context.Context, state model.FsmState, journey model.Journey[T], data any) (model.Journey[T], any, string, *nuErrors.Error) {
	log := logging.GetLogger(ctx)
	resp, updatedJourneyData, nextEvent, err := state.StateHandler.Visit(ctx, journey.JID, journey.Data, data)
	if err != nil {
		log.Errorf("State handler visit method failed with error: %+v", err)
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
	log := logging.GetLogger(ctx)
	resp, updatedJourneyData, err := state.StateHandler.Revisit(ctx, journey.JID, journey.Data)
	if err != nil {
		log.Errorf("State handler revisit method failed with error: %+v", err)
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

func (fs fsmService[T]) startNewJourney(ctx context.Context, data any) (model.Journey[T], any, string, *nuErrors.Error) {
	log := logging.GetLogger(ctx)
	initState, err := fs.getState(ctx, fs.initialStateName)
	if err != nil {
		return model.Journey[T]{}, nil, "", err
	}
	journey, err := fs.journeyStore.Create(ctx)
	if err != nil {
		log.Errorf("Error from journey store. Error: %+v", err)
		return model.Journey[T]{}, nil, "", err
	}
	jid := journey.JID
	journey, resp, nextEvent, err := fs.handleStateVisit(ctx, initState, journey, data)
	if err != nil {
		log.Info("Rolling back journey creation")
		deleteErr := fs.journeyStore.Delete(ctx, jid)
		if deleteErr != nil {
			log.Warnf("Error from journey store. Error: %+v", err)
		}
		return model.Journey[T]{}, nil, "", err
	}

	return journey, resp, nextEvent, nil
}

func (fs fsmService[T]) revisitAndSave(ctx context.Context, journey model.Journey[T], state model.FsmState) (model.FsmResponse, *nuErrors.Error) {
	log := logging.GetLogger(ctx)
	journey, resp, err := fs.handleStateRevisit(ctx, state, journey)
	if err != nil {
		return model.FsmResponse{}, err
	}
	err = fs.journeyStore.Save(ctx, journey)
	if err != nil {
		log.Errorf("Error from journey store. Error: %+v", err)
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
