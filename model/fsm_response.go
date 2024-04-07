package model

type FsmResponse struct {
	JID  string      `json:"jID"`
	Data interface{} `json:"data"`
}
