package model

type FsmResponse struct {
	JID        string `json:"jID"`
	Data       any    `json:"data,omitempty"`
	NextScreen string `json:"next_screen,omitempty"`
	MetaData   any    `json:"meta_data,omitempty"`
}
