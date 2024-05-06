package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type LoadBalancer struct {
	replicas []string
	mux      sync.Mutex
}

var lb LoadBalancer

type ConsistentHashMap struct {
	numSlots      int
	slots         []int
	numContainers int
	numVirtuals   int
}

func NewConsistentHashMap(numSlots, numContainers, numVirtuals int) *ConsistentHashMap {
	c := &ConsistentHashMap{
		numSlots:      numSlots,
		slots:         make([]int, numSlots),
		numContainers: numContainers,
		numVirtuals:   numVirtuals,
	}
	for i := range c.slots {
		c.slots[i] = -1 // Initialize slots with -1 indicating empty
	}
	c.setupVirtualServers()
	return c
}

func (c *ConsistentHashMap) setupVirtualServers() {
	for i := 0; i < c.numContainers; i++ {
		for j := 0; j < c.numVirtuals; j++ {
			slot := c.virtualServerHash(i, j) % c.numSlots

			// Linear probing for collision resolution
			for c.slots[slot] != -1 {
				slot = (slot + 1) % c.numSlots
			}

			c.slots[slot] = i
		}
	}
}

func (c *ConsistentHashMap) virtualServerHash(serverID, virtualID int) int {
	return serverID + virtualID + 2*virtualID + 25
}

func (c *ConsistentHashMap) requestHash(requestID int) int {
	return (requestID*requestID + 2*requestID + 217) % c.numSlots
}

func (c *ConsistentHashMap) mapRequest(requestID int) int {
	slot := c.requestHash(requestID)
	startSlot := slot
	for c.slots[slot] == -1 {
		slot = (slot + 1) % c.numSlots
		if slot == startSlot {
			// Ensure we do not loop indefinitely
			panic("No available server found")
		}
	}
	return c.slots[slot]
}

func main() {
	lb = LoadBalancer{replicas: []string{"Server 1", "Server 2", "Server 3"}}

	http.HandleFunc("/rep", handleReplicas)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/rm", handleRemove)
	http.HandleFunc("/", handleRouting)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func handleReplicas(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": map[string]interface{}{
			"N":        len(lb.replicas),
			"replicas": lb.replicas,
		},
		"status": "successful",
	})
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

}

func handleRouting(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

}
