package balancer

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neerajjain92/loadbalancer-in-golang/internal/server"
)

type ServerPool struct {
	backends []*server.Server
	current  uint64
	mutex    sync.Mutex
}

type Config struct {
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
	Weights             []int    `json:"weights"`
	ListentPort         string   `json:"listenPort"`
}

func LoadConfig(file string) (Config, error) {
	var config Config
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	return config, err
}

func NewServerPool(backends []*server.Server, weight []int) *ServerPool {
	return &ServerPool{backends: backends}
}

func (s *ServerPool) NextBackendRoundRobin() *server.Server {
	next := atomic.AddUint64(&s.current, 1)
	return s.backends[next%(uint64(len(s.backends)))] // Rounding it off with total servers
}

// Other LoadBalancing Algorithms

func (s *ServerPool) LoadBalancer(w http.ResponseWriter, r *http.Request, algorithm string) {
	var backend *server.Server
	switch algorithm {
	case "roundrobin":
		backend = s.NextBackendRoundRobin()
		// ... (other algorithm)
	default:
		http.Error(w, "Invalid Load Balancing Algorithm", http.StatusBadRequest)
		return
	}

	if backend != nil && backend.Alive {
		backend.Server.Handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Service Not available", http.StatusServiceUnavailable)
	}
}
