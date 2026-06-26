package workload

import (
	"sync/atomic"
)

type Tracker struct {
	activeJobs atomic.Int32
}

func NewTracker() *Tracker {
	return &Tracker{}
}

func (t *Tracker) ActiveJobs() int {
	return int(t.activeJobs.Load())
}

func (t *Tracker) Increment() {
	t.activeJobs.Add(1)
}

func (t *Tracker) Decrement() {
	t.activeJobs.Add(-1)
}
