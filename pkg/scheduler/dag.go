package scheduler


type TaskStatus int
const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

type DAGTask struct{

	ID int
	title string
	Goal string
	DependsOn []int
	Status TaskStatus
	Output string
	Err error
}

type DAG struct{

	tasks map[int]*DAGTask
}