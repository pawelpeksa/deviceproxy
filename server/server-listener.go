package server

// Listener ...
type Listener interface {
	OnConnectionEstabilishedFromClient(connectionID string, deviceTag *string) error
	OnMessageReceivedFromClient(connectionID string, msg *string, deviceTag *string) error
	OnClientDisconnected(connectionID string, deviceTag *string) error
	OnServerStopped()
}
