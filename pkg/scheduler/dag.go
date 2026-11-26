package scheduler

import (
	"errors"
	"fmt"
)


type TaskStatus int
const (
	StatusPending TaskStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

type DAGTask struct{

	Id int
	Title string
	Goal string
	DependsOn []int
	Status TaskStatus
	Output string
	Err error
}

type DAG struct{

	tasks map[int]*DAGTask
}
//constructor 
func NewDAG()*DAG{
	return &DAG{
		tasks: make(map[int]*DAGTask),
	}
}

func(d *DAG) AddTask(task *DAGTask){
	if task == nil {
		return
	}
	d.tasks[task.Id] = task
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

///dispactcher

type TaskResult struct {
	TaskId int
	WorkerId int
	Output  string
	Err  error
}


func (d *DAG) Execute(queue chan<- *DAGTask,results <-chan TaskResult)error{


	totalTasks := len(d.tasks)

	completedCount := 0

	for _, task := range d.GetReadyTasks(){
		task.Status = StatusRunning
		fmt.Println("sending task to queue: ",task.Id)
		queue <- task
	}

	for completedCount < totalTasks{

		res := <-results

		task := d.tasks[res.TaskId]

		if res.Err != nil{
			task.Status = StatusFailed
			task.Err = res.Err

			return errors.New("abort the DAG")
		}

		task.Status = StatusCompleted
		task.Output = res.Output
		completedCount++

		fmt.Println("res: ",res.Output)

		for _, nextTask := range d.GetReadyTasks(){
			nextTask.Status = StatusRunning
			queue <- nextTask
		}
	}

	return nil

}

