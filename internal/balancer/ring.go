package balancer

import (
	"log"
	"net/url"
	"sort"
)

type Ring struct {
	servers map[uint32]string
	keys    []uint32
}

func NewRing() *Ring {
	return &Ring{
		servers: make(map[uint32]string),
		keys:    make([]uint32, 0),
	}
}

func (ring *Ring) AddServer(server string) {
	parsedURL, err := url.Parse(server)
	if err != nil {
		log.Fatalf("Failed to parse server URL: %v", err)
	}

	hostWithPort := parsedURL.Host
	hash := hashServer(hostWithPort)
	ring.servers[hash] = hostWithPort
	ring.keys = append(ring.keys, hash)
	sort.Slice(ring.keys, func(i, j int) bool { return ring.keys[i] < ring.keys[j] })
}

func (ring *Ring) RemoveServer(server string) {
	parsedURL, err := url.Parse(server)
	if err != nil {
		log.Fatalf("Failed to parse server URL: %v", err)
	}
	hostWithPort := parsedURL.Host
	hash := hashServer(hostWithPort)
	delete(ring.servers, hash) // Delete from map

	// In the slice you should skip that element and join the prefix slice and the suffix slice
	for index, key := range ring.keys {
		if key == hash {
			ring.keys = append(ring.keys[:index], ring.keys[index+1:]...)
			break
		}
	}
}

func (ring *Ring) GetServer(key string) string {
	hash := hashServer(key)
	index := sort.Search(len(ring.keys), func(i int) bool { return ring.keys[i] >= hash })
	if index == len(ring.keys) {
		index = 0
	}
	return ring.servers[ring.keys[index]]
}
