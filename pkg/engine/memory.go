package engine

import (
	"errors"
	"slices"
	"strings"
)

//token estimater ,priority rentention, and sliding window,


type Message struct{

	Role string
	Content string
	TokenCount int
	Pinned bool
}

type ContextWindow struct {

	maxTokens int
	messages []Message
	currentToken int
}
func EstimateToken(text string)(int){


	return (len(text)/4) + 4
}


func NewContextWindow(maxTokens int) *ContextWindow{

	return &ContextWindow{
		maxTokens: maxTokens,
		messages: make([]Message,0),
		currentToken:0,
	}
}

func (cw *ContextWindow)evictIfNeeded() bool{

	//if there is not enough space
	for cw.currentToken > cw.maxTokens {
		
		targetIndex :=-1

		for i,m := range cw.messages{

			if !m.Pinned{
				targetIndex = i
				break
			}
		}

		if targetIndex == -1 {
			//no unpinned msg
			return false 
		}

		cw.currentToken = cw.currentToken - cw.messages[targetIndex].TokenCount
		
		cw.messages = slices.Delete(cw.messages,targetIndex,targetIndex+1)

		
	}

	return true

}

func (cw *ContextWindow)Add(role string,content string,pinned bool) (string,error){


	if !pinned && EstimateToken(content) > (cw.maxTokens/2){
		
		maxChars := (cw.maxTokens/2)*4

		if len(content) > maxChars {
			content = content[:maxChars] + "\n...[LONG MESSAGE TRUNCATED]..."
		}
	}
	tokens:= EstimateToken(content)

	cw.messages=append(cw.messages,Message{
		Role: role,
		Content: content,
		TokenCount: tokens,
		Pinned: pinned,
	})
	
	cw.currentToken += tokens

	if !cw.evictIfNeeded(){
		return "",errors.New("Context window full and no unpinned messages to remove")
	}

	return "Ok",nil
}


func (cw *ContextWindow)RenderPrompt()string{

	var sb strings.Builder

	for _,msg := range cw.messages{

		sb.WriteString(msg.Content)
		sb.WriteString("\n\n")
	}

	return sb.String()
}
