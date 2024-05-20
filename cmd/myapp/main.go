package main

import (
	"encoding/json"
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

	// Initialize the consistentHashRing
	ring := balancer.NewRing()
	for _, serverUrl := range config.Servers {
		ring.AddServer(serverUrl)
	}

	// Create a new request multiplexer
	mux := http.NewServeMux()

	serverPool := balancer.NewServerPool(backendServers, config.Weights, ring)
	// Register the load balancer handler for all paths
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serverPool.LoadBalancer(w, r, "consistentHashing")
	})

	// Endpoint to add a server
	mux.HandleFunc("/add-server", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ServerURL string `json:"server_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request to add a new server", http.StatusBadRequest)
			return
		}
		serverPool.AddServer(req.ServerURL)
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/remove-server", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ServerURL string `json:"server_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid Request to remove a server", http.StatusBadRequest)
			return
		}
		serverPool.RemoveServer(req.ServerURL)
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("LoadBalancer started at :%s", config.ListentPort)
	if err := http.ListenAndServe(":"+config.ListentPort, mux); err != nil {
		log.Fatalf("Could not start load balancer :%s\n", err)
	}
}
