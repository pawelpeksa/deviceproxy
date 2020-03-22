package api_test

import (
	"testing"

	"github.com/bxcodec/faker"
	"github.com/stretchr/testify/suite"

	queuewrapper "git.krk.awesome-ind.com/GoUtils/QueueWrapper"

	"deviceproxy/api"
	"deviceproxy/mocks_test"
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

	msg string
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

	faker.FakeData(&suite.msg)
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

	err := suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.msg, &suite.publishQueueTopic0)
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

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.msg).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.msg, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.msg).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.msg, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnClientDisconnected(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.msg, &suite.deviceTag0)
	suite.NotNil(err)
}

func (suite *ServerTestSuite) Test_APIPublishMessageToQueueIfMessageComesFromClient() {
	err := suite.api.OnConnectionEstabilishedFromClient(suite.connectionID0, &suite.deviceTag0)
	suite.Nil(err)

	suite.msgQueueMock.On("PublishMessage", suite.publishQueueTopic0, suite.msg).Return(nil)
	err = suite.api.OnMessageReceivedFromClient(suite.connectionID0, &suite.msg, &suite.deviceTag0)
	suite.Nil(err)
}
