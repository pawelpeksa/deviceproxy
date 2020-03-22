package model

// ResponseMsg ...
type ResponseMsg struct {
	Code    int    `json:"code" bson:"code"`
	Message string `json:"message" bson:"message"`
}
