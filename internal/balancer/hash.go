package balancer

import "hash/fnv"

func hashServer(server string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(server))
	return hash.Sum32()
}
