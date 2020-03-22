package api

import (
	"fmt"
	"log"
	"strings"
	"sync"

	queuewrapper "git.krk.awesome-ind.com/GoUtils/QueueWrapper"

	"deviceproxy/resources"
)

const (
	logPrefix = "DeviceProxyAPI "
)

//NOTE: access to data of this structure has to be synchronized
// API ...
type API struct {
	msgSender          MessageSender
	sessionsPerClients map[string][]string
	clientsPerTopics   map[string]int
	msgQueue           queuewrapper.IMsgQueue
	connectionMutex    sync.RWMutex
	messageMutex       sync.RWMutex
}

// NewAPI ...
func NewAPI(msgSender MessageSender, msgQueue queuewrapper.IMsgQueue) *API {
	return &API{
		msgSender:          msgSender,
		clientsPerTopics:   map[string]int{},
		connectionMutex:    sync.RWMutex{},
		messageMutex:       sync.RWMutex{},
		msgQueue:           msgQueue,
		sessionsPerClients: make(map[string][]string),
	}
}

// OnConnectionEstabilishedFromClient ...
func (api *API) OnConnectionEstabilishedFromClient(connectionID string, deviceTag *string) error {
	api.connectionMutex.Lock()
	defer api.connectionMutex.Unlock()

	eventQueueTopic := getEventQueueTopic(deviceTag)

	clientsPerTopics, msgQueueExist := api.clientsPerTopics[eventQueueTopic]

	if msgQueueExist {
		api.clientsPerTopics[eventQueueTopic] = clientsPerTopics + 1
		log.Printf(logPrefix+"OnConnectionEstabilishedFromClient: %v connected clients for topic:%v\n", api.clientsPerTopics[eventQueueTopic], eventQueueTopic)
		return nil
	}

	err := api.msgQueue.AddSubscription(eventQueueTopic, api, true)

	if err != nil {
		log.Printf(logPrefix+"Error while subscribing to topic %v: %v\n", eventQueueTopic, err)
		return err
	}

	api.clientsPerTopics[eventQueueTopic] = 1

	return nil
}

// OnMessageReceivedFromClient ...
func (api *API) OnMessageReceivedFromClient(connectionID string, msg *string, deviceTag *string) error {
	api.messageMutex.Lock()
	defer api.messageMutex.Unlock()

	publishQueueTopic := getPublishQueueTopic(deviceTag)
	eventQueueTopic := getEventQueueTopic(deviceTag)

	if !api.queueExists(eventQueueTopic) {
		return fmt.Errorf(logPrefix + "OnMessageReceivedFromClient: Trying to send message to queue to the topic of the device but service is not registered to listen to it")
	}

	logDebug(fmt.Sprintf("Publishing message %v to queue topic: %v \n", publishQueueTopic, *msg))
	err := api.msgQueue.PublishMessage(publishQueueTopic, *msg)

	if err != nil {
		return fmt.Errorf(logPrefix+"OnMessageReceivedFromClient: Error publishing client message to queue: %v\n", err)
	}

	return nil
}

// OnClientDisconnected ...
func (api *API) OnClientDisconnected(connectionID string, deviceTag *string) error {
	api.connectionMutex.Lock()
	defer api.connectionMutex.Unlock()

	eventQueueTopic := getEventQueueTopic(deviceTag)
	clientsPerTopics, msgQueueExist := api.clientsPerTopics[eventQueueTopic]

	_, ok := api.sessionsPerClients[connectionID]

	if ok {
		delete(api.sessionsPerClients, connectionID)
	}

	if !msgQueueExist {
		log.Printf(logPrefix+"OnClientDisconnected: Could not find queue with topic %v\n", eventQueueTopic)
		return nil
	}

	api.clientsPerTopics[eventQueueTopic] = clientsPerTopics - 1

	if api.clientsPerTopics[eventQueueTopic] > 0 {
		log.Printf(logPrefix+"OnClientDisconnected: %v connected clients left for topic:%v\n", api.clientsPerTopics[eventQueueTopic], eventQueueTopic)
		return nil
	}

	err := api.msgQueue.RemoveSubscription(eventQueueTopic)

	if err != nil {
		log.Printf(logPrefix+"OnClientDisconnected: Could not remove subscription for topic:%v Error:%v\n", eventQueueTopic, err)
		return err
	}

	delete(api.clientsPerTopics, eventQueueTopic)

	log.Printf(logPrefix+"OnClientDisconnected: Subscription for topic:%v has been removed\n", eventQueueTopic)

	return nil
}

// OnServerStopped ...
func (api *API) OnServerStopped() {
	api.msgQueue.ShutDown()
	log.Printf(logPrefix + "OnServerStopped: API closed all queues\n")
}

// ParseMsg ...
func (api *API) ParseMsg(topic, msg string) error {
	if !api.queueExists(topic) {
		log.Printf(logPrefix + "Error: Wrong state of the service. Unscubscribing from topic should have happened\n")
		return nil // we don't want to return err to queue because it'll retry to deliver the message
	}

	logDebug(fmt.Sprintf("Received msg: %v for topic: %v \n", msg, topic))

	deviceTag := getDeviceTagFromTopic(topic)
	return api.msgSender.SendMsg(msg, deviceTag)
}

func (api *API) queueExists(topic string) bool {
	api.connectionMutex.RLock()
	defer api.connectionMutex.RUnlock()

	_, msgQueueExist := api.clientsPerTopics[topic]

	return msgQueueExist
}

func getPublishQueueTopic(deviceTag *string) string {
	return "cloud.msg." + *deviceTag
}

func getEventQueueTopic(deviceTag *string) string {
	return "edge.msg." + *deviceTag
}

func getDeviceTagFromTopic(queueTopic string) string {
	s := strings.Split(queueTopic, ".msg.")
	if len(s) < 2 {
		panic(logPrefix + "This function has been called with wrong argument")
	}

	return s[1]
}

func logDebug(s string) {
	if resources.DeviceProxyLogDebug {
		log.Printf(logPrefix + s)
	}
}
