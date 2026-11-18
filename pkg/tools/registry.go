package tools



// Tool definations, registry map And Context timeouts


import (
	"context"
	"errors"
	"sync"
	"time"
)


type ToolFunc func(ctx context.Context, input string) (string ,error)


type registeredTool struct {

	fn ToolFunc
	cb *CircuitBreaker
}

type Registry struct {

	mutexx sync.RWMutex
	tools map[string] registeredTool
}

func NewRegistry() *Registry{


	return &Registry{
		mutexx: sync.RWMutex{},
		tools: make(map[string]registeredTool),
	}
}

func (r * Registry) Register(name string, fn ToolFunc){


	r.mutexx.Lock()
	defer r.mutexx.Unlock()

	r.tools[name] = registeredTool{
		fn: fn,
		cb: NewCircuitBreaker(5, 100*time.Millisecond),
	}

}

func (r *Registry)Execute(ctx context.Context,name string , input string, timeout time.Duration) (string, error){


	r.mutexx.RLock()

	tool,exists := r.tools[name]
	if! exists {
		r.mutexx.RUnlock()
		return "",errors.New("tool not found: "+name)
	}
	r.mutexx.RUnlock()

	tCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()


	return tool.cb.Execute(func()(string,error){
		return tool.fn(tCtx,input)
	})

}