package port

type WorkloadTracker interface {
	ActiveJobs() int
	Increment()
	Decrement()
}
