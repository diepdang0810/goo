package domain

type StateMachine interface {
	CanTransition(from Status, to Status) bool
}
