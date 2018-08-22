package task

import (
	"time"

	"../currency"
)

// State : state of task
type State int

const (
	StateFree      State = 0
	StateExecuting State = 1
	StatePaused    State = 2
	StateCompleted State = 3
	StateAccepted  State = 4
	StateClosed    State = 5
)

// Task : structure for task
type Task struct {
	ID            int
	CustomerID    int
	ExecutionerID int
	Title         string
	State         State
	Cost          currency.Money
	Problem       string
	Solution      string
	BeginTime     time.Time
	EndTime       time.Time
}
