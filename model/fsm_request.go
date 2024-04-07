package model

type FsmRequest struct {
	JID   string      `json:"jID"`
	Event string      `json:"event" binding:"required"`
	Data  interface{} `json:"data"`
}
