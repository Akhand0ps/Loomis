package main

import (
	"agent-engine/pkg/engine"
	"agent-engine/pkg/tools"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	// "os"
	"sync"
	// "time"
)

type Task struct {
	Id      int
	Goal    string
}
type TaskResult struct{
	TaskId   int
	WorkerId  int
	Output    string
	Err      error
}

func worker(reg *tools.Registry, workerid int, queue <-chan Task, results chan<- TaskResult) {

	for job := range queue {
		// fmt.Printf("worker %d started for jobid : %d | payload: %s\n", workerid, job.Id, job.Payload)
		// time.Sleep(time.Second)
		// fmt.Println()
		// fmt.Printf("worker %d finished for the jobid : %d\n", workerid, job.Id)

		ctx := context.Background()
		// output,err := reg.Execute(ctx,job.ToolName,job.Payload,2*time.Second)
		// if err != nil {
		// 	fmt.Printf("worker %d: tool error: %v\n",workerid,err)
		// } else {
		// 	fmt.Printf("worker %d: tool output: %s\n",workerid,output)
		// }


		mock:= &MockLLM{}
		agent := engine.NewAgent(mock,reg,5)
		ans,err := agent.Run(ctx,job.Goal)
		if err != nil{
			results <- TaskResult{
				TaskId:job.Id,
				WorkerId:workerid,
				Output:"",
				Err: err,
			}
		}else{
			results <- TaskResult{
				TaskId:job.Id,
				WorkerId:workerid,
				Output:ans,
				Err: nil,
			}
		}
	}
}


//using chat <-int because only sender can close the channel not the receiver.
func closeWorker(wg *sync.WaitGroup,results chan<- TaskResult){
	wg.Wait()  //pause here until worker finish
	close(results) //shut down the pipe


}

type MockLLM struct{

	step int
}

func(m *MockLLM)Generate(ctx context.Context,prompt string)(string, error){

	if m.step == 0{
		m.step++
		return `{
			"thought":"I need to check the weather in Tokyo. ",
			"action":"tool_call",
			"tool_name":"weather",
			"tool_input":"Tokyo"
		
		}`,nil
	}

	return fmt.Sprintf(`{
		"though":"I have the weather information now.",
		"action":"final_answer",
		"final_answer":"The weather in %s is 18 degrees celcius and clear skies!."
	}`,"tokyo"),nil

}


type ChatMessage struct{

	Role  string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct{
	Model string `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ResponseContent struct{

	Content string  `json:"content"`
}

type ResponseMessage struct {
 
	Message ResponseContent  `json:"message"`
}
type ChatResponse struct{
	Choices []ResponseMessage `json:"choices"`

}

type OpenRouterClient struct{

	apiKey string
	model  string
}


func(c * OpenRouterClient) Generate(ctx context.Context, prompt string)(string,error){

	reqBody := ChatRequest{

		Model: c.model,
		
		Messages: []ChatMessage{
			{
				Role: "user",
				Content: prompt,
			},
		},
	}

	rawBody,err := json.Marshal(reqBody) //to get json bytes

	if err != nil{
		return "",err
	}

	request ,err := http.NewRequestWithContext(ctx,"POST","https://openrouter.ai/api/v1/chat/completions",bytes.NewBuffer(rawBody)) //create req with body


	request.Header.Set("Authorization","Bearer " + c.apiKey)
	request.Header.Set("Content-Type","application/json")

	//send via client 
	client := &http.Client{}
	resp,err := client.Do(request)
	if err != nil{
		return "",err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		errBody,_ := io.ReadAll(resp.Body)
		return "",fmt.Errorf("openrouter error : status code: %d err:%s",resp.StatusCode,string(errBody))
	}

	// var chatResp ChatResponse
	// if err := json.NewDecoder(resp.Body).Decode(&chatResp);err != nil{

	// 	return "",err
	// }

	var chatResp ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil{
		return "",err
	}

	if len(chatResp.Choices) == 0 {
		return "",fmt.Errorf("empty response from openrouter")
	}


	return chatResp.Choices[0].Message.Content , nil


}
func main() {

	reg := tools.NewRegistry()

	reg.Register("flaky",func(ctx context.Context,input string)(string,error){
		return "",fmt.Errorf("external 500 error for input: %s",input)
	})

	reg.Register("weather",func(ctx context.Context, input string) (string, error) {
		return fmt.Sprintf("18 degree celcius, clear skies in %s",input),nil
	})

	// apiKey := os.Getenv("OPENROUTER_API_KEY")
	// if apiKey == ""{
	// 	fmt.Println("Please set OPENROUTER_API_KEY")
	// 	return
	// }

	// llm := &OpenRouterClient{
	// 	apiKey: apiKey,
	// 	model: "google/gemini-2.0-flash-exp:free",
	// }

	// agent := engine.NewAgent(llm,reg,5)

	// ans,err := agent.Run(context.Background(),"What is the weather in Paris rightnow?")
	// if err != nil {
	// 	fmt.Println("error: ",err)
	// } else {
		// fmt.Println("answer: ",ans)
	// }

	taskcnt := 5
	queue := make(chan Task, taskcnt) //buffered
	results := make(chan TaskResult, taskcnt) //buffered

	var wg sync.WaitGroup


	for i := 1; i <= 3; i++ {
		wg.Go(func() {
			worker(reg,i, queue, results)

		})
	}

	for i := 1; i <= 5; i++ {

		switch i{
		
		case 1:
			task := Task{
				Id: i,
				Goal: "Tokyo",
			}
			queue <- task
			
		case 2:
			task := Task{
				Id: i,
				Goal: "London",
			}
			queue <- task
		case 3:
			task := Task{
				Id: i,
				Goal: "Paris",
			}
			queue <- task
		case 4:
			task := Task{
				Id: i,
				Goal: "Jalandhar",
			}
			queue <- task
		case 5:
			task := Task{
				Id: i,
				Goal: "Delhi",
			}
			queue <- task

		}


	}
	close(queue)

	go closeWorker(&wg,results);

	for result := range results {

		// fmt.Println("execution_done: ", result)

		if result.Err != nil{
			fmt.Printf("Worker %d failed task %d: %v\n",result.WorkerId,result.TaskId,result.Err)
		} else{
			fmt.Printf("Worker %d finished task %d: %s\n",result.WorkerId,result.TaskId,result.Output)
		}

	}


	
}
