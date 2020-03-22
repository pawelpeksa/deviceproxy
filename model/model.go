package model

// ActionMsg ...
type ActionMsg struct {
	Action string `json:"action" bson:"action"`
}

// ResponseMsg ...
type ResponseMsg struct {
	Code    int    `json:"code" bson:"code"`
	Message string `json:"message" bson:"message"`
}

// ActionMsg ...
type BasicActionMsg struct {
	Action    string `json:"action" bson:"action"`
	Timestamp int64  `json:"timestamp" bson:"timestamp"`
	ID        string `json:"id" bson:"id"`
}
