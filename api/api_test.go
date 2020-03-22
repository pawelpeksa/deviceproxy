package api_test

import (
	queuewrapper "git.krk.awesome-ind.com/GoUtils/QueueWrapper"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.azc.ext.hp.com/fitstation-isaac/appinapp_service/internal/deviceproxy/api"
	"github.azc.ext.hp.com/fitstation-isaac/appinapp_service/internal/deviceproxy/mocks_test"
	"github.azc.ext.hp.com/fitstation-isaac/appinapp_service/internal/deviceproxy/model"

	"encoding/json"
	"testing"
)

type ServerTestSuite struct {
	suite.Suite

	api               *api.API
	messageSenderMock *mocks_test.MessageSender
	msgQueueMock      *queuewrapper.MsgQueueMock

	connectionID0 string
	connectionID1 string

	deviceTag0         string
	eventQueueTopic0   string
	publishQueueTopic0 string

	deviceTag1         string
	eventQueueTopic1   string
	publishQueueTopic1 string

	eventMsgJSON string

	actionMsg     model.ActionMsg
	actionMsgJSON string

	emptyActionMsg     model.ActionMsg
	emptyActionMsgJSON string

	openSessionActionMsgJSON  string
	openSessionActionMsgJSON2 string
	openSessionActionMsgJSON3 string

	closeSessionActionMsg      model.BasicActionMsg
	closeSessionActionMsgJSON  string
	closeSessionActionMsg2     model.BasicActionMsg
	closeSessionActionMsgJSON2 string
	closeSessionActionMsg3     model.BasicActionMsg
	closeSessionActionMsgJSON3 string
}

func TestExecuteServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) SetupTest() {
	suite.messageSenderMock = &mocks_test.MessageSender{}
	suite.msgQueueMock = queuewrapper.NewMsgQueueMock()

	suite.api = api.NewAPI(suite.messageSenderMock, suite.msgQueueMock)

	suite.connectionID0 = "piesek"
	suite.connectionID1 = "leszek"

	faker.FakeData(&suite.deviceTag0)
	suite.eventQueueTopic0 = "edge.msg." + suite.deviceTag0
	suite.publishQueueTopic0 = "cloud.msg." + suite.deviceTag0

	faker.FakeData(&suite.deviceTag1)
	suite.eventQueueTopic1 = "edge.msg." + suite.deviceTag1
	suite.publishQueueTopic1 = "cloud.msg." + suite.deviceTag1

	suite.eventMsgJSON = "{}"

	faker.FakeData(&suite.actionMsg)
	actionMsgJSON, err := json.Marshal(&suite.actionMsg)
	suite.Nil(err)

	suite.actionMsgJSON = string(actionMsgJSON)

	faker.FakeData(&suite.emptyActionMsg)
	suite.emptyActionMsg.Action = ""

	emptyActionMsgJSON, err := json.Marshal(&suite.emptyActionMsg)
	suite.Nil(err)

	suite.emptyActionMsgJSON = string(emptyActionMsgJSON)

	// open 1
	bytes, err := json.Marshal(model.BasicActionMsg{
		Action:    api.ActionRSScanOpenSession,
		Timestamp: 0,
		ID:        "grzesio",
	})
	suite.openSessionActionMsgJSON = string(bytes)
	// open 2
	bytes, err = json.Marshal(model.BasicActionMsg{
		Action:    api.ActionRSScanOpenSession,
		Timestamp: 0,
		ID:        "krzesio",
	})
	suite.openSessionActionMsgJSON2 = string(bytes)
	// open 3
	bytes, err = json.Marshal(model.BasicActionMsg{
		Action:    api.ActionRSScanOpenSession,
		Timestamp: 0,
		ID:        "lesio",
	})
	suite.openSessionActionMsgJSON3 = string(bytes)
	// close 1
	suite.closeSessionActionMsg = model.BasicActionMsg{
		Action:    api.ActionRSScanCloseSession,
		Timestamp: 0,
		ID:        "grzesio",
	}
	bytes, err = json.Marshal(suite.closeSessionActionMsg)
	suite.closeSessionActionMsgJSON = string(bytes)
	// close 2
	suite.closeSessionActionMsg2 = model.BasicActionMsg{
		Action:    api.ActionRSScanCloseSession,
		Timestamp: 0,
		ID:        "krzesio",
	}
	bytes, err = json.Marshal(suite.closeSessionActionMsg2)
	suite.closeSessionActionMsgJSON2 = string(bytes)
	// close 3
	suite.closeSessionActionMsg3 = model.BasicActionMsg{
		Action:    api.ActionRSScanCloseSession,
		Timestamp: 0,
		ID:        "lesio",
	}
	bytes, err = json.Marshal(suite.closeSessionActionMsg3)
	suite.closeSessionActionMsgJSON3 = string(bytes)
}

func (suite *ServerTestSuite) AfterTest(suiteName, testName string) {
	suite.messageSenderMock.AssertExpectations(suite.T())
	suite.msgQueueMock.AssertExpectations(suite.T())
}

func (suite *ServerTestSuite) Test_MessageIsSentToClientIfItComesFromQueue() {
	suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)

	suite.messageSenderMock.On("SendMsg", suite.eventMsgJSON, suite.deviceTag0).Once().Return(nil)
	queuewrapper.SendQueueMsgToService(suite.msgQueueMock, suite.eventQueueTopic0, suite.eventMsgJSON)
}

func (suite *ServerTestSuite) Test_APIIsUnsubcribedFromTopicAfterClientIsDisconnected() {
	suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.api.OnConnectionEstabilishedFromClient(suite.connectionID1, &suite.deviceTag1)

	suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)

	err := suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.actionMsgJSON, &suite.publishQueueTopic0)
	suite.NotNil(err)

	queuewrapper.SendQueueMsgToService(suite.msgQueueMock, suite.eventQueueTopic0, suite.eventMsgJSON)

	suite.messageSenderMock.On("SendMsg", suite.eventMsgJSON, suite.deviceTag1).Once().Return(nil)
	queuewrapper.SendQueueMsgToService(suite.msgQueueMock, suite.eventQueueTopic1, suite.eventMsgJSON)
}

func (suite *ServerTestSuite) Test_APIReturnsNoErrorIfTwoClientsWantsToSpeakWithTheSameDevice() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnConnectionEstabilishedFromClient(suite.connectionID1, &suite.deviceTag0)
	suite.Nil(err)
}

func (suite *ServerTestSuite) Test_TopicIsNotUnsubscribedIfThereIsStillClientConnectedToIt() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.actionMsgJSON).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.actionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.actionMsgJSON).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.actionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.actionMsgJSON, &suite.deviceTag0)
	suite.NotNil(err)
}

func (suite *ServerTestSuite) Test_APIPublishMessageToQueueIfMessageComesFromClient() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.actionMsgJSON).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.actionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)
}

func (suite *ServerTestSuite) Test_APIReturnsErrorIfAPISendsMessageWithEmptyAction() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.emptyActionMsgJSON, &suite.deviceTag0)
	suite.NotNil(err)
}

func (suite *ServerTestSuite) Test_APISendsCloseMessagesForStillOpenedSessionAfterConnectionDrops() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.openSessionActionMsgJSON).Return(nil)
	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.openSessionActionMsgJSON2).Return(nil)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.openSessionActionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.openSessionActionMsgJSON2, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.closeSessionActionMsgJSON).Return(nil)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.closeSessionActionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.getActionsComparator(suite.closeSessionActionMsg2)).Return(nil)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)
}

func (suite *ServerTestSuite) Test_APISendsCloseMessagesForStillOpenedSessionAfterConnectionDropsAndWhichAnyOfThemWasNotClosedManually() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.openSessionActionMsgJSON).Return(nil)
	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.openSessionActionMsgJSON2).Return(nil)
	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.openSessionActionMsgJSON3).Return(nil)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.openSessionActionMsgJSON, &suite.deviceTag0)
	suite.Nil(err)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.openSessionActionMsgJSON2, &suite.deviceTag0)
	suite.Nil(err)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.openSessionActionMsgJSON3, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.getActionsComparator(suite.closeSessionActionMsg)).Return(nil)
	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.getActionsComparator(suite.closeSessionActionMsg2)).Return(nil)
	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.getActionsComparator(suite.closeSessionActionMsg3)).Return(nil)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)
}

func (suite *ServerTestSuite) getActionsComparator(compareWith model.BasicActionMsg) interface{} {
	return mock.MatchedBy(func(res interface{}) bool {
		baseActionMessage := &model.BasicActionMsg{}
		err := json.Unmarshal([]byte(res.(string)), baseActionMessage)
		suite.Nil(err)
		return baseActionMessage.Action == compareWith.Action && baseActionMessage.ID == compareWith.ID
	})
}
