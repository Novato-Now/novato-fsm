package model

type Journey[T any] struct {
	JID                 string `json:"jID"`
	CurrentStage        string `json:"current_stage"`
	LastCheckpointStage string `json:"last_checkpoint_stage"`
	Data                T      `json:"data"`
}
