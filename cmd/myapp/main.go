package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/neerajjain92/loadbalancer-in-golang/internal/balancer"
	"github.com/neerajjain92/loadbalancer-in-golang/internal/server"
)

func main() {
	// Load Configuration
	config, err := balancer.LoadConfig("config/config.json")
	if err != nil {
		log.Fatalf("Failed to load config %v", err)
	}

	// Start the backend Servers
	var wg sync.WaitGroup
	backendServers := make([]*server.Server, len(config.Servers))
	for i, serverURL := range config.Servers {
		wg.Add(1)
		backendServers[i] = server.NewServer(serverURL)
		go func(s *server.Server) {
			defer wg.Done()
			s.Start()
		}(backendServers[i])
	}

	// Wait for backend servers to Start
	wg.Wait() // Until all respective wg.Done is being called for each server

	// Initialize ServerPool
	healthChecker, err := balancer.NewHealthChecker(backendServers, config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Failed to initialize health checker: %v", err)
	}
	go healthChecker.Start()

	// Create a new request multiplexer

	mux := http.NewServeMux()

	serverPool := balancer.NewServerPool(backendServers, config.Weights)
	// Register the load balancer handler for all paths
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverPool.LoadBalancer(w, r, "roundrobin")
	})

	log.Printf("LoadBalancer started at :%s", config.ListentPort)
	if err := http.ListenAndServe(":"+config.ListentPort, mux); err != nil {
		log.Fatalf("Could not start load balancer :%s\n", err)
	}
}
