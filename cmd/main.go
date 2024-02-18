package main

import (
	"microservice/http/orchestrator_http"
	"microservice/internal/agent"
	"microservice/pkg/database"
	"sync"
)

func main() {
	database.Create()
	var goroutines int = 3
	var wg sync.WaitGroup
	wg.Add(1)
	orchestrator_http.RunOrchestrator()
	agent.RunAgent(goroutines)
	wg.Wait()
}
