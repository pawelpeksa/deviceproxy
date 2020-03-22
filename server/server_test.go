package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/bxcodec/faker"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"

	"deviceproxy/mocks_test"
	"deviceproxy/model"
	"deviceproxy/server"
)

type ServerTestSuite struct {
	suite.Suite

	listenerMock *mocks_test.Listener
	server       *server.Server

	connection        *websocket.Conn
	testConnectionURL string

	connectionID1 string
	connectionID2 string
	deviceTag1    string
	deviceTag2    string
}

func TestExecuteServerTestSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

func (suite *ServerTestSuite) SetupTest() {
	suite.listenerMock = &mocks_test.Listener{}

	suite.server = server.NewServer()
	suite.server.Listener = suite.listenerMock

	testServer := httptest.NewServer(http.HandlerFunc(suite.server.ProxyHandler))

	suite.testConnectionURL = "ws" + strings.TrimPrefix(testServer.URL, "http")

	faker.FakeData(&suite.deviceTag1)
	faker.FakeData(&suite.deviceTag2)
	faker.FakeData(&suite.connectionID1)
	faker.FakeData(&suite.connectionID2)
}

func (suite *ServerTestSuite) AfterTest(suiteName, testName string) {
	suite.listenerMock.AssertExpectations(suite.T())
}

func (suite *ServerTestSuite) Test_ConnectionIsDroppedIfDeviceTagIsNotProvided() {
	clientConnection, _, err := websocket.DefaultDialer.Dial(suite.testConnectionURL, nil)
	defer clientConnection.Close()

	suite.Nil(err)

	responseMsg := model.ResponseMsg{}
	clientConnection.ReadJSON(&responseMsg)
	suite.EqualValues(-1, responseMsg.Code)
}

func (suite *ServerTestSuite) Test_ConnectionIsEstabilishedIfDeviceTagIsProvided() {
	suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
}

func (suite *ServerTestSuite) Test_ListenerIsCalledWhenMessageIsSent() {
	clientConnection1 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	defer clientConnection1.Close()

	clientConnection2 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag2)
	defer clientConnection2.Close()

	msg := "msg"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg, &suite.deviceTag1).Once().Return(nil)

	err := clientConnection1.WriteMessage(websocket.TextMessage, []byte(msg))
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection1)
}

func (suite *ServerTestSuite) Test_MultipleConnectionsCanSendMessageToOneDeviceTag() {
	clientConnection1 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	defer clientConnection1.Close()

	clientConnection2 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	defer clientConnection2.Close()

	//client1
	msg1 := "msg1"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg1, &suite.deviceTag1).Once().Return(nil)

	err := clientConnection1.WriteMessage(websocket.TextMessage, []byte(msg1))
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection1)

	//client2
	msg2 := "msg2"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg2, &suite.deviceTag1).Once().Return(nil)

	err = clientConnection2.WriteMessage(websocket.TextMessage, []byte(msg2))
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection2)
}

func (suite *ServerTestSuite) Test_SecondConnectionSendsSuccesfullyToClientIfTheFirstOneDrops() {
	clientConnection1 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	clientConnection2 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)

	msg1 := "msg1"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg1, &suite.deviceTag1).Once().Return(nil)

	err := clientConnection1.WriteMessage(websocket.TextMessage, []byte(msg1))
	suite.Nil(err)
	suite.expectSuccesfullResponse(clientConnection1)

	suite.listenerMock.On("OnClientDisconnected", mock.Anything, &suite.deviceTag1).Times(1).Return(nil)
	clientConnection1.Close()

	msg2 := "msg2"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg2, &suite.deviceTag1).Once().Return(nil)

	err = clientConnection2.WriteMessage(websocket.TextMessage, []byte(msg2))
	suite.Nil(err)
	suite.expectSuccesfullResponse(clientConnection2)

	err = suite.server.SendMsg("{}", suite.deviceTag1)
	suite.Nil(err)

	suite.listenerMock.On("OnClientDisconnected", mock.Anything, &suite.deviceTag1).Times(1).Return(nil)
	clientConnection2.Close()
	time.Sleep(50 * time.Millisecond)
}

func (suite *ServerTestSuite) Test_ClientIsNotDisconnectedWhenOtherOneDropsWithTheSameDeviceTag() {
	clientConnection1 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)

	clientConnection2 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	defer clientConnection2.Close()

	//client1
	msg1 := "msg1"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg1, &suite.deviceTag1).Once().Return(nil)

	err := clientConnection1.WriteMessage(websocket.TextMessage, []byte(msg1))
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection1)

	suite.listenerMock.On("OnClientDisconnected", mock.Anything, &suite.deviceTag1).Times(1).Return(nil)

	clientConnection1.Close()

	//client2
	msg2 := "msg2"

	suite.listenerMock.On("OnMessageReceivedFromClient", mock.Anything, &msg2, &suite.deviceTag1).Once().Return(nil)

	err = clientConnection2.WriteMessage(websocket.TextMessage, []byte(msg2))
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection2)
}

func (suite *ServerTestSuite) Test_ListenerIsCalledWhenClientIsDisconnected() {
	clientConnection1 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag1)
	defer clientConnection1.Close()

	clientConnection2 := suite.estabilishClientConnectionForDeviceTag(suite.deviceTag2)
	defer clientConnection2.Close()

	suite.listenerMock.On("OnClientDisconnected", mock.Anything, &suite.deviceTag1).Once().Return(nil)

	clientConnection1.Close()
	time.Sleep(50 * time.Millisecond)
}

func (suite *ServerTestSuite) expectSuccesfullResponse(clientConnection *websocket.Conn) {
	responseMsg := model.ResponseMsg{}
	err := clientConnection.ReadJSON(&responseMsg)
	suite.Nil(err)
	suite.EqualValues(0, responseMsg.Code)
}

func (suite *ServerTestSuite) estabilishClientConnectionForDeviceTag(deviceTag string) *websocket.Conn {
	suite.listenerMock.On("OnConnectionEstabilishedFromClient", mock.Anything, &deviceTag).Once().Return(nil)

	testConnectionURL := appendDeviceTagToURL(suite.testConnectionURL, deviceTag)

	clientConnection, _, err := websocket.DefaultDialer.Dial(testConnectionURL, nil)
	suite.Nil(err)

	suite.expectSuccesfullResponse(clientConnection)

	return clientConnection
}

func appendDeviceTagToURL(URL, deviceTag string) string {
	return fmt.Sprintf("%s/?deviceTag=%s", URL, deviceTag)
}
