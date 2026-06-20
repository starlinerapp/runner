package port

type WorkloadTracker interface {
	ActiveJobs() int
}
