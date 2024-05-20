package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type Server struct {
	URL    *url.URL
	Server *http.Server
	Mutex  sync.RWMutex
	Alive  bool
}

func NewServer(serverURL string) *Server {
	url, _ := url.Parse(serverURL)
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request on Server %v, Method: %s, URL: %s\n", serverURL, r.Method, r.URL.Path)
		fmt.Fprintf(w, "Hello from server %s!", url.Host)
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// log.Printf("Received HealthCheck on Server %v \n", serverURL)
		fmt.Fprintf(w, "Server [%s] is healthy !!", url.Host)
	})

	return &Server{
		URL:   url,
		Alive: true,
		Server: &http.Server{
			Addr:    url.Host,
			Handler: mux,
		},
	}
}

func (s *Server) Start() {
	log.Printf("Starting server on %s \n", s.URL.Host)
	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s : err==>>%s \n", s.URL.Host, err)
		}
	}()
}

func (s *Server) Stop() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	if s.Server != nil {
		log.Printf("Stopping Server on %s \n", s.URL.Host)
		s.Server.Close()
		s.Alive = false
	}
}
