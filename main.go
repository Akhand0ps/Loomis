package main

import (
	"fmt"
	"sync"
	"time"

)

type Task struct {
	Id      int
	Payload string
}

func worker(workerid int, queue <-chan Task, results chan<- int) {

	for job := range queue {
		fmt.Printf("worker %d started for jobid : %d | payload: %s\n", workerid, job.Id, job.Payload)
		time.Sleep(time.Second)
		fmt.Println()
		fmt.Printf("worker %d finished for the jobid : %d\n", workerid, job.Id)
		results <- job.Id
	}
}


//using chat <-int because only sender can close the channel not the receiver.
func closeWorker(wg *sync.WaitGroup,results chan<- int){
	wg.Wait()  //pause here until worker finish
	close(results) //shut down the pipe


}

func main() {

	taskcnt := 10
	queue := make(chan Task, taskcnt)
	results := make(chan int, taskcnt)

	var wg sync.WaitGroup

	payload := "akhand"

	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			worker(i, queue, results)

		})
	}

	for i := 1; i <= 5; i++ {

		payload = fmt.Sprintf("akhand%d", i)

		task := Task{Id: i, Payload: payload}

		queue <- task

	}
	close(queue)

	go closeWorker(&wg,results);

	for result := range results {

		fmt.Println("Results ", result)
	}



}
