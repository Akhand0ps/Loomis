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

func NewDAG()*DAG{
	return &DAG{
		tasks: make(map[int]*DAGTask),
	}
}

func(d *DAG) AddTask(task *DAGTask){
	if task == nil {
		return
	}
	d.tasks[task.ID] = task
}

func(d *DAG) IsReadyToRun(task *DAGTask)bool{

	if task.Status != StatusPending {
		return false
	}

	for _,parentId := range task.DependsOn{
		parent,exists := d.tasks[parentId]
		if !exists || parent.Status != StatusCompleted {
			return false
		}
	}

	return  true
}

func(d * DAG) GetReadyTasks()[]*DAGTask{

	var readyTasks []*DAGTask


	for _,task := range d.tasks{
		if d.IsReadyToRun(task){
			readyTasks = append(readyTasks,task)
		}
	}

	return readyTasks
}