# Loomis: A Custom Event-Driven Agent Orchestration Engine (From Scratch)

## Project Overview
A production-grade, highly concurrent agent orchestration engine built completely from scratch in Go, avoiding high-level frameworks (like LangChain or CrewAI). The engine accepts complex, multi-step user intents, compiles them into a Directed Acyclic Graph (DAG) of dependent sub-tasks, executes those tasks concurrently using specialized sub-agents, manages LLM context tokens via a sliding memory window, and handles network/API failures gracefully using circuit breakers.

## Core Architectural Components
1. **The Orchestration Core (Scheduler & DAG Engine):** Parses instructions and coordinates the execution flow of dependent or parallel tasks.
2. **The ReAct Execution Loop:** A strict state machine that manages the "Reasoning + Acting" cycle of individual agents using raw LLM APIs.
3. **The Async Tool Runner:** A sandboxed execution pool that runs external APIs or local scripts concurrently with strict timeout and retry policies.
4. **The Context Window & Memory Gate:** A token-aware memory manager that dynamically compacts or truncates history to prevent context overflow.

## Technical Signals (For Interviewers)
* **Concurrency Primitives:** Heavy utilization of native Go Goroutines, Channels, and Mutexes for safe data routing.
* **Resilience Patterns:** Custom implementation of Circuit Breakers and Exponential Backoff for distributed API interactions.
* **Memory & I/O Optimization:** Low-level token calculation, text parsing, and non-blocking network streams.
