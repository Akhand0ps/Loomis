package main

import (
	"agent-engine/pkg/tools"
	"context"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	Id      int
	ToolName string
	Payload string
}

func worker(reg *tools.Registry, workerid int, queue <-chan Task, results chan<- int) {

	for job := range queue {
		// fmt.Printf("worker %d started for jobid : %d | payload: %s\n", workerid, job.Id, job.Payload)
		// time.Sleep(time.Second)
		// fmt.Println()
		// fmt.Printf("worker %d finished for the jobid : %d\n", workerid, job.Id)

		ctx := context.Background()
		output,err := reg.Execute(ctx,job.ToolName,job.Payload,2*time.Second)
		if err != nil {
			fmt.Printf("worker %d: tool error: %v\n",workerid,err)
		} else {
			fmt.Printf("worker %d: tool output: %s\n",workerid,output)
		}

		results <- job.Id
	}
}


//using chat <-int because only sender can close the channel not the receiver.
func closeWorker(wg *sync.WaitGroup,results chan<- int){
	wg.Wait()  //pause here until worker finish
	close(results) //shut down the pipe


}

func main() {

	reg := tools.NewRegistry()

	reg.Register("flaky",func(ctx context.Context,input string)(string,error){
		return "",fmt.Errorf("external 500 error for input: %s",input)
	})

	taskcnt := 10
	queue := make(chan Task, taskcnt) //buffered
	results := make(chan int, taskcnt) //buffered

	var wg sync.WaitGroup

	payload := "akhand"

	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			worker(reg,i, queue, results)

		})
	}

	for i := 1; i <= 8; i++ {

		payload = fmt.Sprintf("akhand%d", i)

		task := Task{Id: i, Payload: payload,ToolName: "flaky"}

		queue <- task

	}
	close(queue)

	go closeWorker(&wg,results);

	for result := range results {

		fmt.Println("execution_done: ", result)
	}


	
}
