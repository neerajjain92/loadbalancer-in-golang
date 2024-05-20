package balancer

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/neerajjain92/loadbalancer-in-golang/internal/server"
)

type ServerPool struct {
	backends []*server.Server
	current  uint64
	ring     *Ring
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

func NewServerPool(backends []*server.Server, weight []int, ring *Ring) *ServerPool {
	return &ServerPool{backends: backends, ring: ring}
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
	case "consistentHashing":
		serverAddr := s.ring.GetServer(r.RemoteAddr)
		for _, server := range s.backends {
			if server.URL.Host == serverAddr {
				backend = server
				break
			}
		}
	default:
		http.Error(w, "Invalid Load Balancing Algorithm", http.StatusBadRequest)
		return
	}

	if backend != nil && backend.Alive {
		// Establish a TCP connection with the backend server
		backendConn, err := net.Dial("tcp", backend.URL.Host)
		if err != nil {
			http.Error(w, "Failed to connect to backend server", http.StatusServiceUnavailable)
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// Send the request to the backend Server in HTTP Format
		go func() {
			defer wg.Done()
			err := r.Write(backendConn)
			if err != nil {
				log.Printf("Error sending request to backend server: %v", err)
			}
		}()

		// Read response from the backend server and write to the client
		go func() {
			defer wg.Done()
			resp, err := http.ReadResponse(bufio.NewReader(backendConn), r)
			if err != nil {
				log.Printf("Error reading response from backend server: %v", err)
				http.Error(w, "Error reading response from backend server", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()
			copyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			backendConn.Close() // Close the backend connection after the response is read
		}()

		wg.Wait()
	} else {
		http.Error(w, "Service not available", http.StatusServiceUnavailable)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vals := range src {
		for _, val := range vals {
			dst.Add(k, val)
		}
	}
}

func (s *ServerPool) AddServer(serverURL string) {
	s.ring.AddServer(serverURL)
	newServer := server.NewServer(serverURL)
	s.backends = append(s.backends, newServer)
	go newServer.Start()
}

func (s *ServerPool) RemoveServer(serverURL string) {
	s.ring.RemoveServer(serverURL)
	for i, backend := range s.backends {
		if backend.URL.String() == serverURL {
			backend.Stop()
			s.backends = append(s.backends[:i], s.backends[i+1:]...)
			break
		}
	}
}
