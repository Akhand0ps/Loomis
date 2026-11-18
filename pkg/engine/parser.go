package engine

import (
	"encoding/json"
	"errors"
	"strings"
)

//A simple parser that converts raw text from an LLM
// into a strongly typed agent command.


type AgentStep struct {

	Thought string `json:"thought"`
	Action string  `json:"action"`
	ToolName string `json:"tool_name"`
	ToolInput string `json:"tool_input"`
	FinalAnswer string `json:"final_answer"`

}

var ResponseError = errors.New("no valid JSON Object found in model output")

func ParseStep(rawResponse string) (AgentStep,error){

	
	
	start := strings.Index(rawResponse, "{")  //index of the first {, or -1 if missing
	end := strings.LastIndex(rawResponse, "}")  


	if start == -1 || end  == -1  || start > end{

		return AgentStep{},ResponseError
	}

	cleanJSON := rawResponse[start:end+1]
	var step AgentStep

	if err := json.Unmarshal([]byte(cleanJSON) , &step); err != nil{

		return AgentStep{},err
	}

	if step.Action != "tool_call" && step.Action != "final_answer" {
		return AgentStep{},errors.New("invalid action: must be 'tool_call' or 'final_answer'")
	}

	if step.Action == "tool_call" && step.ToolName == "" {
		return AgentStep{},errors.New("tool_name cannot be empty for tool_call action")
	}

	return step,nil

	
}