package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"sync"
	"time"
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
	lb = LoadBalancer{replicas: []string{"server_1", "server_2", "server_3"}}

	http.HandleFunc("/rep", handleReplicas)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/rm", handleRemove)
	http.HandleFunc("/", handleRouting)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func handleReplicas(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	w.Header().Set("Content-Type", "application/json")
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

	var payload struct {
		N         int      `json:"n"`
		Hostnames []string `json:"hostnames"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(payload.Hostnames) > payload.N {
		http.Error(w, "Length of hostname list is more than newly added instances", http.StatusBadRequest)
		return
	}

	for i := len(payload.Hostnames); i < payload.N; i++ {
		payload.Hostnames = append(payload.Hostnames, generateRandomHostname())
	}

	for _, hostname := range payload.Hostnames {
		cmd := exec.Command("docker", "run", "-d", "--name", hostname, "-e", fmt.Sprintf("SERVER_ID=%s", hostname), "dslb-server")
		if err := cmd.Run(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to start Docker container: %v", err), http.StatusInternalServerError)
			return
		}
		lb.replicas = append(lb.replicas, hostname)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": map[string]interface{}{
			"N":        len(lb.replicas),
			"replicas": lb.replicas,
		},
		"status": "successful",
	})
}

func generateRandomHostname() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return "server_" + string(b)
}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

	var payload struct {
		N         int      `json:"n"`
		Hostnames []string `json:"hostnames"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(payload.Hostnames) > payload.N {
		http.Error(w, "Length of hostname list is more than removable instances", http.StatusBadRequest)
		return
	}

	// Randomly select additional instances for removal if necessary
	toRemove := payload.Hostnames
	if len(toRemove) < payload.N {
		existingReplicas := lb.replicas
		rand.Seed(time.Now().UnixNano())
		for len(toRemove) < payload.N {
			randomIndex := rand.Intn(len(existingReplicas))
			randomInstance := existingReplicas[randomIndex]
			if !contains(toRemove, randomInstance) {
				toRemove = append(toRemove, randomInstance)
			}
		}
	}

	// Remove instances from the replicas list and stop the Docker containers
	for _, hostname := range toRemove {
		if err := stopAndRemoveDockerContainer(hostname); err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove Docker container: %v", err), http.StatusInternalServerError)
			return
		}
		lb.replicas = removeFromSlice(lb.replicas, hostname)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": map[string]interface{}{
			"N":        len(lb.replicas),
			"replicas": lb.replicas,
		},
		"status": "successful",
	})
}

func stopAndRemoveDockerContainer(containerName string) error {
	cmd := exec.Command("docker", "stop", containerName)
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("docker", "rm", containerName)
	return cmd.Run()
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func handleRouting(w http.ResponseWriter, r *http.Request) {
	lb.mux.Lock()
	defer lb.mux.Unlock()

}
