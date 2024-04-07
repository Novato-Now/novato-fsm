package model

import "fsm/action"

type FsmState struct {
	Name                string
	Action              action.Action
	NextAvailableEvents []NextAvailableEvent
	IsCheckpoint        bool
}

type NextAvailableEvent struct {
	Event                string
	DestinationStateName string
}
