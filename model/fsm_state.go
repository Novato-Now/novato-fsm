package model

import "github.com/Novato-Now/novato-fsm/state_handler"

type FsmState struct {
	Name                string
	StateHandler        state_handler.StateHandler
	NextAvailableEvents []NextAvailableEvent
	IsCheckpoint        bool
	NextScreen          string
	MetaData            any
}

type NextAvailableEvent struct {
	Event                string
	DestinationStateName string
}
