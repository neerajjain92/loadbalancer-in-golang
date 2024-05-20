package balancer

import (
	"net/http"
	"time"

	"github.com/neerajjain92/loadbalancer-in-golang/internal/server"
)

type HealthChecker struct {
	servers  []*server.Server
	interval time.Duration
}

func NewHealthChecker(servers []*server.Server, interval string) (*HealthChecker, error) {
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return nil, err
	}
	return &HealthChecker{
		servers:  servers,
		interval: duration,
	}, nil
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	for {
		select {
		case <-ticker.C:
			hc.CheckHealth()
		}
	}
}

func (hc *HealthChecker) CheckHealth() {
	for _, backend := range hc.servers {
		go func(backendServer *server.Server) {
			resp, err := http.Get(backendServer.URL.String() + "/health")
			backendServer.Mutex.Lock()
			defer backendServer.Mutex.Unlock()
			if err != nil || resp.StatusCode != http.StatusOK {
				backendServer.Alive = false
			} else {
				backendServer.Alive = true
			}
		}(backend)
	}
}
