# Loomis: Concurrent Agent Execution Engine (From Scratch in Go)

A production-grade, highly concurrent agent orchestration runtime built completely from scratch in Go, free from high-level agent frameworks (such as LangChain or CrewAI). 

Loomis dispatches user goals across a bounded worker pool, manages multi-turn **Reason + Act (ReAct)** execution loops, guards external APIs with custom **3-State Circuit Breakers**, and prevents cascading failures using strict backpressure and context timeouts.

---

## Architecture Overview

```
                      [ User Goals / Task Ingestion ]
                                     │
                                     ▼
                    [ Bounded Queue Channel (Backpressure) ]
                                     │
            ┌────────────────────────┼────────────────────────┐
            ▼                        ▼                        ▼
      [ Worker 1 ]              [ Worker 2 ]             [ Worker 3 ]
   (Goroutine Worker)        (Goroutine Worker)       (Goroutine Worker)
            │                        │                        │
            └────────────────────────┼────────────────────────┘
                                     │
                                     ▼
                        [ ReAct State Machine ]
                         ┌───────────────────┐
                         │   1. Thought      │ (LLM / MockClient)
                         │   2. Action       │ (tool_call vs final_answer)
                         │   3. Observation  │ (Tool output / error)
                         └─────────┬─────────┘
                                   │
                                   ▼
                        [ Tool Registry Runner ]
                     (sync.RWMutex + Context Timeout)
                                   │
                                   ▼
                    [ 3-State Circuit Breaker ]
                    ┌─────────────────────────┐
                    │ Closed ──► Open ──► Half │
                    └─────────────────────────┘
                                   │
                                   ▼
                       [ Results Channel Stream ]
                                   │
                                   ▼
                  [ Graceful Drainer (sync.WaitGroup) ]
```

---

## Core Engineering Features

### 1. High-Throughput Worker Pool (Fan-Out / Fan-In)
- **Controlled Concurrency:** Workers pull tasks dynamically from a shared, bounded channel (`queue`), preventing CPU thrashing and unbounded goroutine creation.
- **Backpressure by Design:** Bounded buffer capacities force producers to pace ingestion when workers are saturated, preventing Out-Of-Memory (OOM) crashes.
- **Zero-Deadlock Graceful Shutdown:** A dedicated background closer goroutine waits on `sync.WaitGroup` before closing the `results` channel, allowing `main` to stream results cleanly without circular dependencies.

### 2. Thread-Safe 3-State Circuit Breaker (`pkg/tools/breaker.go`)
- **Finite State Machine:** Implements `StateClosed` (healthy), `StateOpen` (tripped/failing fast), and `StateHalfOpen` (probing recovery).
- **Sub-Microsecond Lock Windows:** Mutex locks are held only for state inspection and transition. Locks are explicitly released *before* slow I/O or network calls to guarantee concurrent workers never serialize on mutex contention.

### 3. Isolated Tool Registry (`pkg/tools/registry.go`)
- **Shared Read / Exclusive Write Safety:** Built on `sync.RWMutex` to allow dozens of concurrent workers to look up tools simultaneously without contention.
- **Per-Tool Isolation:** Every tool maintains its own dedicated `CircuitBreaker`. A degraded third-party API will only trip its own breaker without affecting unrelated tools.
- **Context Timeout Enforcement:** Wraps tool executions in `context.WithTimeout`, automatically reclaiming hung sockets and worker threads.

### 4. Self-Healing ReAct Engine (`pkg/engine/`)
- **JSON Sanitizer & Validator (`parser.go`):** Slices JSON between boundary braces (`{ ... }`), eliminating markdown fuzz and conversational chatter. Validates actions and prevents empty tool invocations.
- **Autonomous Self-Correction (`react.go`):** If an LLM returns malformed syntax, the state machine appends the error directly into conversation history and loops forward, allowing the model to repair its own JSON without crashing the workflow.
- **Observation Feeding:** Tool errors and circuit-trip notifications are fed back into the reasoning loop as observations, empowering agents to detect degraded tools and attempt fallback strategies.

---

## Project Structure

```
agent-engine/
├── pkg/
│   ├── engine/
│   │   ├── parser.go       # JSON extraction, sanitization, and step validation
│   │   └── react.go        # Multi-turn ReAct state machine & self-healing loop
│   └── tools/
│       ├── breaker.go      # 3-state Circuit Breaker (Closed, Open, Half-Open)
│       └── registry.go     # Thread-safe tool directory with timeout enforcement
├── main.go                 # Orchestrator entry point, worker pool & mock simulation
└── go.mod                  # Go module definition (Go 1.26+)
```

---

## Getting Started

### Prerequisites
- Go 1.22+ (utilizes per-iteration loop variable semantics and modern concurrency primitives)

### Running the Orchestrator
Clone the repository and run the engine:

```bash
go run main.go
```

### Example Output
```text
history:  You are an autonomous agent that solves tasks using a Reason + Act loop...
User Goal: Tokyo

history:  You are an autonomous agent that solves tasks using a Reason + Act loop...
User Goal: London

Worker 1 finished task 1: The weather in Tokyo is 18 degrees celcius and clear skies!.
Worker 1 finished task 4: The weather in Tokyo is 18 degrees celcius and clear skies!.
Worker 1 finished task 5: The weather in Tokyo is 18 degrees celcius and clear skies!.
Worker 2 finished task 2: The weather in Tokyo is 18 degrees celcius and clear skies!.
Worker 3 finished task 3: The weather in Tokyo is 18 degrees celcius and clear skies!.
```

---

## Technical Highlights for Systems & Backend Interviews
- **Go Runtime Internals:** Explores the GMP scheduler, `hchan` circular ring buffers, and goroutine memory characteristics (~2KB stack growth).
- **Concurrency Correctness:** Clean prevention of goroutine leaks, data races (verified via `go test -race`), and channel deadlocks.
- **Zero Framework Bloat:** Demonstrates that robust agent orchestration requires only foundational systems architecture: worker pools, state machines, and resilience patterns.
