package model

type FsmResponse struct {
	JID        string      `json:"jID"`
	Data       interface{} `json:"data,omitempty"`
	NextScreen string      `json:"next_screen,omitempty"`
	MetaData   interface{} `json:"meta_data,omitempty"`
}
