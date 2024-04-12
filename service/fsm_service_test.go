package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	fsmErrors "github.com/thevibegod/fsm/errors"
	"github.com/thevibegod/fsm/mocks"
	"github.com/thevibegod/fsm/model"
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
	suite.ctx = context.Background()
}

func (suite *fsmServiceTestSuite) TestNewFsmService_ShouldReturnError_WhenNoFinalStateIsFound() {
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
			Name:                "StateB",
			StateHandler:        suite.mockStateHandler,
			NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)

	suite.Empty(service)
	suite.Equal(fsmErrors.InternalSystemError("no final state found"), err)
}

func (suite *fsmServiceTestSuite) TestNewFsmService_ShouldReturnError_WhenMultipleFinalStatesAreFound() {
	initState := model.FsmState{
		Name:                "Init",
		StateHandler:        suite.mockStateHandler,
		IsCheckpoint:        true,
		NextAvailableEvents: []model.NextAvailableEvent{{Event: "Next", DestinationStateName: "StateA"}},
	}
	nonInitStates := []model.FsmState{
		{
			Name:         "StateA",
			StateHandler: suite.mockStateHandler,
		},
		{
			Name:         "StateB",
			StateHandler: suite.mockStateHandler,
		},
	}

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)

	suite.Empty(service)
	suite.Equal(fsmErrors.InternalSystemError("multiple final states found"), err)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
	suite.Nil(err)

	response, err := service.Execute(suite.ctx, model.FsmRequest{Event: "StartNew"})

	suite.Empty(response)
	suite.Equal(fsmErrors.ByPassError("invalid journey error: wrong event"), err)
}

func (suite *fsmServiceTestSuite) TestExecute_ShouldReturnError_WhenUserStartsNewJourneyAndStoreCreateFails() {
	expectedError := fsmErrors.InternalSystemError("some-error")

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	suite.Equal(fsmErrors.ByPassError("invalid event NextEvent for state Init"), err)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("cannot find next state")

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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	expectedError := fsmErrors.InternalSystemError("some-error")
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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

	service, err := NewFsmService(initState, nonInitStates, suite.mockJourneyStore)
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
	suite.Equal(fsmErrors.InternalSystemError("cannot find next state"), err)
}
