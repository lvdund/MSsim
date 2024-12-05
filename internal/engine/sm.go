package engine

type UeStateMachine interface {
}

type SimpleUeSM struct {
	deregistrationGap, idleGap, reconnectGap, xnHoGap, n2HoGap int
}
