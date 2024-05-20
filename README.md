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

- Now for testing consistent hashing there are endpoint exposed to add or remove servers while loadBalancer is running

```
curl -X POST -H "Content-Type: application/json" -d '{"server_url":"http://localhost:7070"}' http://localhost:9095/remove-server
```

And to add new server
```
curl -X POST -H "Content-Type: application/json" -d '{"server_url":"http://localhost:7070"}' http://localhost:9095/add-server
```