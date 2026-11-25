package engine

import (
	"agent-engine/pkg/tools"
	"context"
	"errors"
	"fmt"

	// "fmt"
	"time"
)


type LLMClient interface {

	Generate(ctx context.Context, prompt string) (string,error)
}

type ReActAgent struct {
	llm LLMClient
	register *tools.Registry
	maxSteps int	
}

func NewAgent(llm LLMClient,register *tools.Registry, maxSteps int) *ReActAgent{

	return &ReActAgent{
		llm: llm,
		register: register,
		maxSteps: maxSteps,	
	}
}

const systemPrompt = `You are an autonomous agent that solves tasks using a Reason + Act loop.
You MUST reply with a single valid JSON object in this exact schema:

{
  "thought": "your step-by-step reasoning",
  "action": "tool_call" OR "final_answer",
  "tool_name": "name of the tool to use (only if action is tool_call)",
  "tool_input": "input to the tool (only if action is tool_call)",
  "final_answer": "your final response to the user (only if action is final_answer)"
}

Rules:
1. Always output ONLY the JSON object. No markdown wrappers or extra conversation.
2. If you have enough information to answer the user's goal, set action to "final_answer".
3. If you need more information, set action to "tool_call" and specify the tool_name and tool_input.`

func (a *ReActAgent) Run(ctx context.Context,goal string)(string,error){

	// history := fmt.Sprintf("%s\n\nUser Goal: %s\n",systemPrompt,goal)
	mem := NewContextWindow(2000) 

	mem.Add("system",fmt.Sprintf("System: %s",systemPrompt),true)
	mem.Add("user",fmt.Sprintf("User Goal: %s",goal),true)

	// fmt.Println("history: ",history)

	for step :=0;step < a.maxSteps; step++{

		raw,err := a.llm.Generate(ctx,mem.RenderPrompt())

		if err != nil{

			return  "",err
		}

		// history += fmt.Sprintf("\nAgent Response:\n%s",raw)
		mem.Add("agent",fmt.Sprintf("agentResponse: %s",raw),false)
		

		parsed,err := ParseStep(raw)

		if err != nil{

			// history += fmt.Sprintf("\nSystem Error: %s",err)
			mem.Add("observation",fmt.Sprintf("observation: %s",err.Error()),false)
			// fmt.Println("parse Error:",err.Error())

			continue
		}

		if parsed.Action == "final_answer"{
			return parsed.FinalAnswer,nil
		}

		if parsed.Action == "tool_call"{
			output,err := a.register.Execute(ctx,parsed.ToolName,parsed.ToolInput,5 *time.Second)

			if err !=nil{
				///
				// history +=fmt.Sprintf("\nTool Error: %s\n",err)
				mem.Add("observation:",fmt.Sprintf("Tool Error: %s",err.Error()),false)
				// continued
			} else{

				// history += fmt.Sprintf("\nObservation: %s\n",output)
				mem.Add("observation:",fmt.Sprintf("Observation: %s",output),false)
			}


		}
	}

	return "",errors.New("max resoning steps exceeded without reaching a final answer")

}