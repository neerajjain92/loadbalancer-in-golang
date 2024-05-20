# QUICK START

## Run the Load Balancer
```
go run cmd/myapp/main.go
```

Configs available at config/config.json
```
{
    "healthCheckInterval": "5s",
    "servers": [
        "http://localhost:7070",
        "http://localhost:7071",
        "http://localhost:7072"
    ],
    "weights": [5,3,2],
    "listenPort": "9095",
    "routingAlgo": "consistentHashing"
}
```

## Client instead of connecting directly to servers connect via load balancer

```
curl localhost:9095

Response:
Hello from server localhost:7070!
```

## Hitting Health Endpoint
```
curl localhost:9095/health
```

## ConsistentHashing

### Remove Server on the fly [Notice how only a handful of request routed to new server rest all remain there, basic concept of consistent hashing]
```
curl -X POST -H "Content-Type: application/json" -d '{"server_url":"http://localhost:7070"}' http://localhost:9095/remove-server
```

### Add Server on the fly
```
curl -X POST -H "Content-Type: application/json" -d '{"server_url":"http://localhost:7070"}' http://localhost:9095/add-server
```