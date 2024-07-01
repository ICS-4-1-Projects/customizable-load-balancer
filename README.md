# Dockerized Load Balancer

## Overview

This project aims to explore the effectiveness of a custom load balancing solution that distributes incoming requests across multiple server containers. The architecture consists of three main components:

- **Server Containers**: Multiple instances of a simple web server running on different ports.
- **Load Balancer**: A Go-based load balancer that uses a consistent hashing algorithm to distribute requests.
- **Chaos Engine**: Utilizes Pumba to simulate network failures and container faults to test the resilience of the load balancer.

## Installation and Running Guide

Follow these steps to deploy and run the system:

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/dockerized-load-balancer.git
   cd dockerized-load-balancer
   ```
2. **Build the Docker Containers**

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
