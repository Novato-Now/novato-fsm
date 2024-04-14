package journeystore

import (
	"context"
	"errors"
	"testing"

	novato_errors "github.com/Novato-Now/novato-utils/errors"
	"github.com/stretchr/testify/suite"
	fsmErrors "github.com/thevibegod/fsm/errors"
	"github.com/thevibegod/fsm/mocks"
	"github.com/thevibegod/fsm/model"
	"go.uber.org/mock/gomock"
)

type testJourneyData struct{}

type journeyStoreTestSuite struct {
	suite.Suite
	mockCtrl          *gomock.Controller
	mockKeyValueStore *mocks.MockKeyValueStore[testJourneyData]
	journeyStore      JourneyStore[testJourneyData]
	ctx               context.Context
}

func TestJourneyStoreTestSuite(t *testing.T) {
	suite.Run(t, new(journeyStoreTestSuite))
}

func (suite *journeyStoreTestSuite) SetupTest() {
	suite.mockCtrl = gomock.NewController(suite.T())
	suite.mockKeyValueStore = mocks.NewMockKeyValueStore[testJourneyData](suite.mockCtrl)
	suite.ctx = context.Background()
	uuidNewString = func() string {
		return "new-uuid"
	}

	suite.journeyStore = NewJourneyStore(suite.mockKeyValueStore)
}

func (suite *journeyStoreTestSuite) TestCreate_ShouldReturnNoError_WhenKeyValueStoreReturnsNoError() {

	expectedJourney := model.Journey[testJourneyData]{JID: "new-uuid"}

	suite.mockKeyValueStore.EXPECT().
		Set(suite.ctx, "FSM_JOURNEY_new-uuid", expectedJourney).
		Return(nil).
		Times(1)

	journey, err := suite.journeyStore.Create(suite.ctx)

	suite.Equal(expectedJourney, journey)
	suite.Nil(err)
}

func (suite *journeyStoreTestSuite) TestCreate_ShouldReturnError_WhenKeyValueStoreReturnsError() {

	expectedJourney := model.Journey[testJourneyData]{JID: "new-uuid"}

	suite.mockKeyValueStore.EXPECT().
		Set(suite.ctx, "FSM_JOURNEY_new-uuid", expectedJourney).
		Return(errors.New("some-error")).
		Times(1)

	journey, err := suite.journeyStore.Create(suite.ctx)

	suite.Empty(journey)
	suite.Equal(novato_errors.InternalSystemError(suite.ctx).WithMessage("some-error"), err)
}

func (suite *journeyStoreTestSuite) TestGet_ShouldReturnNoError_WhenKeyValueStoreReturnsNoError() {

	expectedJourney := model.Journey[testJourneyData]{JID: "new-uuid"}

	suite.mockKeyValueStore.EXPECT().
		Get(suite.ctx, "FSM_JOURNEY_new-uuid").
		Return(&expectedJourney, nil).
		Times(1)

	journey, err := suite.journeyStore.Get(suite.ctx, "new-uuid")

	suite.Equal(expectedJourney, journey)
	suite.Nil(err)
}

func (suite *journeyStoreTestSuite) TestGet_ShouldReturnError_WhenKeyValueStoreReturnsNoJourney() {
	suite.mockKeyValueStore.EXPECT().
		Get(suite.ctx, "FSM_JOURNEY_new-uuid").
		Return(nil, nil).
		Times(1)

	journey, err := suite.journeyStore.Get(suite.ctx, "new-uuid")

	suite.Empty(journey)
	suite.Equal(fsmErrors.BypassError().WithMessage("journey not found"), err)
}

func (suite *journeyStoreTestSuite) TestGet_ShouldReturnError_WhenKeyValueStoreReturnsError() {
	suite.mockKeyValueStore.EXPECT().
		Get(suite.ctx, "FSM_JOURNEY_new-uuid").
		Return(nil, errors.New("some-error")).
		Times(1)

	journey, err := suite.journeyStore.Get(suite.ctx, "new-uuid")

	suite.Empty(journey)
	suite.Equal(novato_errors.InternalSystemError(suite.ctx).WithMessage("some-error"), err)
}

func (suite *journeyStoreTestSuite) TestSave_ShouldReturnNoError_WhenKeyValueStoreReturnsNoError() {

	journey := model.Journey[testJourneyData]{JID: "new-uuid"}

	suite.mockKeyValueStore.EXPECT().
		Set(suite.ctx, "FSM_JOURNEY_new-uuid", journey).
		Return(nil).
		Times(1)

	err := suite.journeyStore.Save(suite.ctx, journey)

	suite.Nil(err)
}

func (suite *journeyStoreTestSuite) TestSave_ShouldReturnError_WhenKeyValueStoreReturnsError() {
	journey := model.Journey[testJourneyData]{JID: "new-uuid"}

	suite.mockKeyValueStore.EXPECT().
		Set(suite.ctx, "FSM_JOURNEY_new-uuid", journey).
		Return(errors.New("some-error")).
		Times(1)

	err := suite.journeyStore.Save(suite.ctx, journey)

	suite.Equal(novato_errors.InternalSystemError(suite.ctx).WithMessage("some-error"), err)
}

func (suite *journeyStoreTestSuite) TestDelete_ShouldReturnNoError_WhenKeyValueStoreReturnsNoError() {

	suite.mockKeyValueStore.EXPECT().
		Del(suite.ctx, "FSM_JOURNEY_new-uuid").
		Return(nil).
		Times(1)

	err := suite.journeyStore.Delete(suite.ctx, "new-uuid")

	suite.Nil(err)
}

func (suite *journeyStoreTestSuite) TestDelete_ShouldReturnError_WhenKeyValueStoreReturnsError() {
	suite.mockKeyValueStore.EXPECT().
		Del(suite.ctx, "FSM_JOURNEY_new-uuid").
		Return(errors.New("some-error")).
		Times(1)

	err := suite.journeyStore.Delete(suite.ctx, "new-uuid")

	suite.Equal(novato_errors.InternalSystemError(suite.ctx).WithMessage("some-error"), err)
}
