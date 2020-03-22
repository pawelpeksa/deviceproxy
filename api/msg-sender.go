package api

// MessageSender ...
type MessageSender interface {
	SendMsg(msg, deviceTag string) error
}
