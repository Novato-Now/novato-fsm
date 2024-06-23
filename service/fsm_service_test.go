package service

import (
	"context"
	"testing"

	fsmErrors "github.com/Novato-Now/novato-fsm/errors"
	"github.com/Novato-Now/novato-fsm/mocks"
	"github.com/Novato-Now/novato-fsm/model"
	"github.com/Novato-Now/novato-utils/constants"
	nuErrors "github.com/Novato-Now/novato-utils/errors"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type testJourneyData struct {
	InitStateCompleted bool
	StateACompleted    bool
	StateBCompleted    bool
}

type fsmServiceTestSuite struct {
	suite.Suite
	mockCtrl         *gomock.Controller
	mockJourneyStore *mocks.MockJourneyStore[testJourneyData]
	mockStateHandler *mocks.MockStateHandler
	ctx              context.Context
}

func TestFsmServiceTestSuite(t *testing.T) {
	suite.Run(t, new(fsmServiceTestSuite))
}

func (suite *fsmServiceTestSuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockStateHandler = mocks.NewMockStateHandler(suite.mockCtrl)
	suite.mockJourneyStore = mocks.NewMockJourneyStore[testJourneyData](suite.mockCtrl)
	suite.ctx = context.WithValue(context.Background(), constants.ServiceNameKey, "FSM")
}

func (suite *fsmServiceTestSuite) TestNewFsmService_ShouldReturnNoError_WhenStatesAreValid() {
	initState := model.FsmState{
		Name:                "Init",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})

	suite.NotEmpty(service)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnNoError_WhenUserStartsNewJourney() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	updatedJourneyData := testJourneyData{InitStateCompleted: true}

	suite.mockJourneyStore.EXPECT().Create(suite.ctx).Return(model.Journey[testJourneyData]{JID: "some-uuid"}, nil).Times(1)
	suite.mockStateHandler.EXPECT().Visit(suite.ctx, "some-uuid", testJourneyData{}, nil).Return(nil, updatedJourneyData, "TransitionComplete", nil).Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, model.Journey[testJourneyData]{JID: "some-uuid", CurrentStage: "Init", LastCheckpointStage: "Init", Data: updatedJourneyData}).
		Return(nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "Start"})

	suite.Equal(model.FsmResponse{JID: "some-uuid", NextScreen: "InitScreen"}, response)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserStartsNewJourney_WithWrongEvent() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "StartNew"})

	suite.Empty(response)
	suite.Equal(fsmErrors.BypassError(), err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserStartsNewJourneyAndStoreCreateFails() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	suite.mockJourneyStore.EXPECT().
		Create(suite.ctx).
		Return(model.Journey[testJourneyData]{}, expectedError).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "Start"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserStartsNewJourneyAndStoreSaveFails() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	updatedJourneyData := testJourneyData{InitStateCompleted: true}

	suite.mockJourneyStore.EXPECT().Create(suite.ctx).Return(model.Journey[testJourneyData]{JID: "some-uuid"}, nil).Times(1)
	suite.mockStateHandler.EXPECT().Visit(suite.ctx, "some-uuid", testJourneyData{}, nil).Return(nil, updatedJourneyData, "TransitionComplete", nil).Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, model.Journey[testJourneyData]{JID: "some-uuid", CurrentStage: "Init", LastCheckpointStage: "Init", Data: updatedJourneyData}).
		Return(expectedError).
		Times(1)
	suite.mockJourneyStore.EXPECT().Delete(suite.ctx, "some-uuid").Return(nil).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "Start"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserStartsNewJourneyAndStateHandlerReturnsError() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	suite.mockJourneyStore.EXPECT().
		Create(suite.ctx).
		Return(model.Journey[testJourneyData]{JID: "some-uuid"}, nil).
		Times(1)

	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", testJourneyData{}, nil).
		Return(nil, nil, "", expectedError).
		Times(1)

	suite.mockJourneyStore.EXPECT().Delete(suite.ctx, "some-uuid").Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "Start"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnNoError_WhenUserTransitionsToStateA() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyData := testJourneyData{InitStateCompleted: true}
	journey := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "Init",
		LastCheckpointStage: "Init",
		Data:                journeyData,
	}
	request := struct{ FieldA bool }{FieldA: true}
	expectedResponse := struct{ ResponseFieldA bool }{ResponseFieldA: true}
	expectedJourneyData := testJourneyData{InitStateCompleted: true, StateACompleted: true}
	expectedJourney := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateA",
		LastCheckpointStage: "Init",
		Data:                expectedJourneyData,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journey, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyData, request).
		Return(expectedResponse, expectedJourneyData, "TransitionComplete", nil).
		Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, expectedJourney).
		Return(nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Next", Data: request})

	suite.Equal(
		model.FsmResponse{
			JID: "some-uuid", NextScreen: "ScreenA", MetaData: "some-metadata", Data: expectedResponse,
		},
		response,
	)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserTransitionsToStateA_AndStoreGetReturnsError() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	request := struct{ FieldA bool }{FieldA: true}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(model.Journey[testJourneyData]{}, expectedError).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Next", Data: request})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserTransitionsToStateA_AndEventNameIsWrong() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyData := testJourneyData{InitStateCompleted: true}
	journey := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "Init",
		LastCheckpointStage: "Init",
		Data:                journeyData,
	}
	request := struct{ FieldA bool }{FieldA: true}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journey, nil).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "NextEvent", Data: request})

	suite.Empty(response)
	suite.Equal(fsmErrors.BypassError(), err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnNoError_WhenUserTransitionsToStateBFromInit() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyDataInit := testJourneyData{InitStateCompleted: true}
	journeyInit := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "Init",
		LastCheckpointStage: "Init",
		Data:                journeyDataInit,
	}
	requestDataA := struct{ FieldA bool }{FieldA: true}
	expectedResponseDataA := struct{ ResponseFieldA bool }{ResponseFieldA: true}
	journeyDataA := testJourneyData{InitStateCompleted: true, StateACompleted: true}
	expectedResponseDataB := struct{ ResponseFieldB bool }{ResponseFieldB: true}
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "Init",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyInit, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyDataInit, requestDataA).
		Return(expectedResponseDataA, journeyDataA, "INTERNAL_Next", nil).
		Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyDataA, expectedResponseDataA).
		Return(expectedResponseDataB, journeyDataB, "TransitionComplete", nil).
		Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, journeyB).
		Return(nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Next", Data: requestDataA})

	suite.Equal(
		model.FsmResponse{
			JID: "some-uuid", NextScreen: "ScreenB", MetaData: "some-metadata", Data: expectedResponseDataB,
		},
		response,
	)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserTransitionsToStateBFromInit_AndStateBVisitReturnsError() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyDataInit := testJourneyData{InitStateCompleted: true}
	journeyInit := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "Init",
		LastCheckpointStage: "Init",
		Data:                journeyDataInit,
	}
	requestDataA := struct{ FieldA bool }{FieldA: true}
	expectedResponseDataA := struct{ ResponseFieldA bool }{ResponseFieldA: true}
	journeyDataA := testJourneyData{InitStateCompleted: true, StateACompleted: true}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyInit, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyDataInit, requestDataA).
		Return(expectedResponseDataA, journeyDataA, "INTERNAL_Next", nil).
		Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyDataA, expectedResponseDataA).
		Return(nil, nil, "", expectedError).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Next", Data: requestDataA})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserTransitionsToStateBFromInit_DueToInvalidStateConfiguration() {
	expectedError := nuErrors.InternalSystemError(suite.ctx)

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateC"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyDataInit := testJourneyData{InitStateCompleted: true}
	journeyInit := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "Init",
		LastCheckpointStage: "Init",
		Data:                journeyDataInit,
	}
	requestDataA := struct{ FieldA bool }{FieldA: true}
	expectedResponseDataA := struct{ ResponseFieldA bool }{ResponseFieldA: true}
	journeyDataA := testJourneyData{InitStateCompleted: true, StateACompleted: true}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyInit, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Visit(suite.ctx, "some-uuid", journeyDataInit, requestDataA).
		Return(expectedResponseDataA, journeyDataA, "INTERNAL_Next", nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Next", Data: requestDataA})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnNoError_WhenUserResumesJourneyInStateB() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
			IsCheckpoint: true,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	expectedResponseDataB := struct{ ResponseFieldB bool }{ResponseFieldB: true}
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Revisit(suite.ctx, "some-uuid", journeyDataB).
		Return(expectedResponseDataB, journeyDataB, nil).
		Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, journeyB).
		Return(nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Resume"})

	suite.Equal(
		model.FsmResponse{
			JID: "some-uuid", NextScreen: "ScreenB", MetaData: "some-metadata", Data: expectedResponseDataB,
		},
		response,
	)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserResumesJourneyInStateB_AndStateHandlerReturnsError() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
			IsCheckpoint: true,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Revisit(suite.ctx, "some-uuid", journeyDataB).
		Return(nil, nil, expectedError).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Resume"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserResumesJourneyInStateB_AndStoreSaveFails() {
	expectedError := nuErrors.InternalSystemError(suite.ctx).WithMessage("some-error")

	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
			IsCheckpoint: true,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	expectedResponseDataB := struct{ ResponseFieldB bool }{ResponseFieldB: true}
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Revisit(suite.ctx, "some-uuid", journeyDataB).
		Return(expectedResponseDataB, journeyDataB, nil).
		Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, journeyB).
		Return(expectedError).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Resume"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserResumesJourney_ForInvalidState() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			MetaData:     "some-metadata",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "State",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Resume"})

	suite.Empty(response)
	suite.Equal(nuErrors.InternalSystemError(suite.ctx), err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnNoError_WhenUserGoesBackToStateAFromStateB() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			IsCheckpoint: true,
			NextAvailableEvents: []model.NextAvailableEvent{
				{Event: "Back", DestinationStateName: "StateA"},
				{Event: "Next", DestinationStateName: "StateC"},
			},
		},
		{
			Name:         "StateC",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenC",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	expectedResponseDataA := struct{ ResponseFieldA bool }{ResponseFieldA: true}
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}
	journeyDataA := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyA := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateA",
		LastCheckpointStage: "StateB",
		Data:                journeyDataA,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)
	suite.mockStateHandler.EXPECT().
		Revisit(suite.ctx, "some-uuid", journeyDataB).
		Return(expectedResponseDataA, journeyDataA, nil).
		Times(1)
	suite.mockJourneyStore.EXPECT().
		Save(suite.ctx, journeyA).
		Return(nil).
		Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Back"})

	suite.Equal(
		model.FsmResponse{
			JID: "some-uuid", NextScreen: "ScreenA", MetaData: "some-metadata", Data: expectedResponseDataA,
		},
		response,
	)
	suite.Nil(err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserGoesBackToStateAFromStateB_AndCurrentStateDoesNotExist() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			IsCheckpoint: true,
			NextAvailableEvents: []model.NextAvailableEvent{
				{Event: "Back", DestinationStateName: "StateA"},
				{Event: "Next", DestinationStateName: "StateC"},
			},
		},
		{
			Name:         "StateC",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenC",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	expectedError := nuErrors.InternalSystemError(suite.ctx)
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateX",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Back"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserGoesBackToStateAFromStateB_AndNextStateDoesNotExist() {
	initState := model.FsmState{
		Name:                "Init",
		NextScreen:          "InitScreen",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:                "StateA",
			StateHandler:        suite.mockStateHandler,
			NextScreen:          "ScreenA",
			MetaData:            "some-metadata",
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "INTERNAL_Next", DestinationStateName: "StateB"}},
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenB",
			IsCheckpoint: true,
			NextAvailableEvents: []model.NextAvailableEvent{
				{Event: "Back", DestinationStateName: "StateX"},
				{Event: "Next", DestinationStateName: "StateC"},
			},
		},
		{
			Name:         "StateC",
			StateHandler: suite.mockStateHandler,
			NextScreen:   "ScreenC",
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore, model.FsmHooks[testJourneyData]{})
	suite.Nil(err)

	expectedError := nuErrors.InternalSystemError(suite.ctx)
	journeyDataB := testJourneyData{InitStateCompleted: true, StateACompleted: true, StateBCompleted: true}
	journeyB := model.Journey[testJourneyData]{
		JID:                 "some-uuid",
		CurrentStage:        "StateB",
		LastCheckpointStage: "StateB",
		Data:                journeyDataB,
	}

	suite.mockJourneyStore.EXPECT().Get(suite.ctx, "some-uuid").Return(journeyB, nil).Times(1)

	response, err := service.Execute(suite.ctx, model.FsmRequest{JID: "some-uuid", Event: "Back"})

	suite.Empty(response)
	suite.Equal(expectedError, err)
}
