# Custom Load Balancer

## Setup Guide

1. Build the server image

```bash
docker build -t dslb-server ./server
```

2. Initialize the cluster

```bash
docker compose up
```

3. Run the load balancer

```bash
go run load-balancer/main.go
```
