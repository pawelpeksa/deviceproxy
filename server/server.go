package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"

	"deviceproxy/model"
	"deviceproxy/resources"
)

const (
	deviceTag = "deviceTag"
	logPrefix = "DeviceProxyServer "
)

// Server ...
type Server struct {
	Listener    Listener
	connections map[string]map[string]*websocket.Conn //NOTE: access to this map has to be synchronized
	httpServer  *http.Server
	mutex       sync.RWMutex
}

// NewServer ...
func NewServer() *Server {
	return &Server{
		connections: map[string]map[string]*websocket.Conn{},
		mutex:       sync.RWMutex{},
	}
}

// Serve ...
func (s *Server) Serve(port string) error {
	http.HandleFunc(resources.DeviceProxyEndpoint, s.ProxyHandler)

	s.httpServer = &http.Server{Addr: ":" + port, Handler: nil}
	err := s.httpServer.ListenAndServe()

	s.Listener.OnServerStopped()

	return err
}

// Shutdown ...
func (s *Server) Shutdown() error {
	return s.httpServer.Shutdown(context.Background())
}

// SendMsg ...
func (s *Server) SendMsg(msg, deviceTag string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	connections, connectionExists := s.connections[deviceTag]

	if !connectionExists {
		return fmt.Errorf(logPrefix+"Can not send message to client because there is not any websocket connection. Device tag:%v", deviceTag)
	}

	for _, connection := range connections {
		err := connection.WriteMessage(websocket.TextMessage, []byte(msg))

		if err != nil {
			log.Printf(logPrefix+"Error sending message to connection. Device tag:%v\n", deviceTag)
		}
	}

	return nil
}

// ProxyHandler ...
func (s *Server) ProxyHandler(wr http.ResponseWriter, req *http.Request) {
	log.Printf(logPrefix+"Incoming request from %s | %s %s\n", req.RemoteAddr, req.Method, req.URL)
	if websocket.IsWebSocketUpgrade(req) {
		s.serveWebSocket(wr, req)
	} else {
		log.Printf(logPrefix + "Incoming request was not socket upgrade\n")
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool {
		return true
	},
}

func (s *Server) serveWebSocket(wr http.ResponseWriter, req *http.Request) {
	connection, err := upgrader.Upgrade(wr, req, nil)
	if err != nil {
		log.Printf(logPrefix+"Error upgrading http request to websocket. Err:%v\n", err)
		return
	}

	defer connection.Close()

	log.Printf(logPrefix+"%v upgraded to websocket\n", req.RemoteAddr)

	sDeviceTag, ok := req.URL.Query()[deviceTag]

	if !ok || len(sDeviceTag) < 1 {
		s.sendResponseMsgToConnection(connection, -1, "URL Param 'deviceTag' is missing")
		return
	}

	deviceTag := sDeviceTag[0]

	if len(deviceTag) < 1 {
		s.sendResponseMsgToConnection(connection, -1, "URL Param 'deviceTag' is missing")
		return
	}

	uuid, _ := uuid.NewV4()
	connectionID := uuid.String()

	s.addConnection(deviceTag, connectionID, connection)

	s.readFromClient(connectionID, connection, deviceTag)

	s.removeConnection(deviceTag, connectionID)
}

func (s *Server) readFromClient(connectionID string, connection *websocket.Conn, deviceTag string) {
	err := s.Listener.OnConnectionEstabilishedFromClient(connectionID, &deviceTag)

	s.sendResponseMsgToClient(connection, deviceTag, 0, "Connection estabilished")

	if err != nil {
		s.sendErrorToClient(connection, deviceTag, fmt.Errorf(logPrefix+"Error OnConnectionEstabilishedFromClient. Err: %v", err))
	}

	for {
		messageType, bMessage, err := connection.ReadMessage()

		if err != nil {
			log.Printf(logPrefix+"Error reading message coming from client. Err: %v", err)
			err = s.Listener.OnClientDisconnected(connectionID, &deviceTag)
			if err != nil {
				log.Printf(logPrefix+"Error on client disconnecting. Err: %v", err)
			}
			break
		}

		if messageType != websocket.TextMessage {
			s.sendErrorToClient(connection, deviceTag, fmt.Errorf(logPrefix+"Incorrect type of message. Client can send only text messages"))
			continue
		}

		message := string(bMessage)

		err = s.Listener.OnMessageReceivedFromClient(connectionID, &message, &deviceTag)

		if err != nil {
			s.sendErrorToClient(connection, deviceTag, err)
			continue
		}

		s.sendResponseMsgToClient(connection, deviceTag, 0, "Message succesfully sent to platform")
	}
}

func (s *Server) sendErrorToClient(connection *websocket.Conn, deviceTag string, err error) {
	s.sendResponseMsgToClient(connection, deviceTag, -1, err.Error())
}

func (s *Server) sendResponseMsgToClient(connection *websocket.Conn, deviceTag string, code int, msg string) {
	s.sendResponseMsgToConnection(connection, code, msg)
}

func (s *Server) sendResponseMsgToConnection(connection *websocket.Conn, code int, msg string) {

	jsonResponse := model.ResponseMsg{
		Code:    code,
		Message: msg,
	}

	sendErr := connection.WriteJSON(jsonResponse)

	if sendErr != nil {
		log.Printf(logPrefix+"Can not send error to client err:%v", sendErr)
		log.Printf(logPrefix+"Message that failed to be sent:%v code:%v", msg, code)
	}
}

func (s *Server) addConnection(deviceTag string, connectionID string, connection *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.connections[deviceTag] == nil {
		s.connections[deviceTag] = make(map[string]*websocket.Conn)
	}

	s.connections[deviceTag][connectionID] = connection
}

func (s *Server) removeConnection(deviceTag string, connectionID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.connections[deviceTag], connectionID)

	if len(s.connections[deviceTag]) == 0 {
		delete(s.connections, deviceTag)
	}
}
