# Custom Load Balancer

## Overview

This project aims to explore the effectiveness of a custom load balancing solution that distributes incoming requests across multiple server containers. The architecture consists of three main components:

- **Server Containers**: Multiple instances of a simple web server running on different ports.
- **Load Balancer**: A Go-based load balancer that uses a consistent hashing algorithm to distribute requests.

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

## Design Choices

- **Consistent Hashing**: Chosen for its efficiency in distributing requests and handling changes in the server pool without significant disruption.
- **Go for Load Balancer**: Offers robust performance and straightforward concurrency, ideal for handling high throughput in network applications.
- **Docker Compose for Orchestration**: Simplifies the deployment and scaling of multiple containers.

## Testing Methodology

- **Performance Testing**: Conducted using Artillery to simulate user traffic and measure response times and error rates.
- **Chaos Experiments**: Pumba was used to introduce network latencies and container crashes to evaluate the load balancer's fault tolerance.

## Challenges and Resolutions

- **Handling Non-uniform Load Distribution**: Early in the testing phase, it was observed that the load was not evenly distributed among all server instances. Some servers were receiving significantly more requests than others, leading to potential bottlenecks and decreased overall system performance.
  - **Resolutions**: Initial consistent hashing algorithms lacked scalability for large clusters or high loads due to a simple linear probing method. To address this, we enhanced the algorithms by increasing the number of virtual nodes for each server, improving load distribution, and boosting system resilience and efficiency.
- **Ensuring Robustness During Network Failures**: During chaos experiments with Pumba, it was found that network delays and disconnections significantly impacted the performance of the load balancer. The system did not gracefully handle the abrupt disconnection of server containers, causing prolonged downtimes and service unavailability.
  - **Resolutions**: Added retry logic in the load balancer to reroute requests to alternative servers if the initially selected server is unresponsive.
