package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type LoadBalancer struct {
	replicas []string
	hashMap  *ConsistentHashMap
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
		// Initialize slots with -1 indicating empty
		c.slots[i] = -1
	}
	c.setupVirtualServers()
	return c
}

func (c *ConsistentHashMap) setupVirtualServers() {
	for i := 0; i < c.numContainers; i++ {
		for j := 0; j < c.numVirtuals; j++ {
			key := fmt.Sprintf("server-%d-virtual-%d", i, j)
			slot := c.hash(key) % c.numSlots

			// Linear probing for collision resolution
			for c.slots[slot] != -1 {
				slot = (slot + 1) % c.numSlots
			}

			c.slots[slot] = i
		}
	}
}

func (c *ConsistentHashMap) hash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}

func (c *ConsistentHashMap) mapRequest(key string) int {
	slot := c.hash(key) % c.numSlots
	startSlot := slot
	for c.slots[slot] == -1 {
		slot = (slot + 1) % c.numSlots
		if slot == startSlot {
			panic("No available server found")
		}
	}
	return c.slots[slot]
}

func (c *ConsistentHashMap) addServer(serverID int, hostname string) {
	for j := 0; j < c.numVirtuals; j++ {
		key := fmt.Sprintf("%s-virtual-%d", hostname, j)
		slot := c.hash(key) % c.numSlots

		// Linear probing for collision resolution
		for c.slots[slot] != -1 {
			slot = (slot + 1) % c.numSlots
		}

		c.slots[slot] = serverID
	}
}

func main() {
	lb = LoadBalancer{
		replicas: []string{"server_1", "server_2", "server_3"},
		hashMap:  NewConsistentHashMap(100, 3, 3),
	}

	for i, hostname := range lb.replicas {
		lb.hashMap.addServer(i, hostname)
	}

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
		port, _ := extractServerPort(hostname)
		cmd := exec.Command("docker", "run", "-d", "--name", hostname, "-e", fmt.Sprintf("SERVER_ID=%s", hostname), "-p", fmt.Sprintf("%d:5000", port), "dslb-server")
		if err := cmd.Run(); err != nil {
			http.Error(w, fmt.Sprintf("Failed to start Docker container: %v", err), http.StatusInternalServerError)
			return
		}
		lb.replicas = append(lb.replicas, hostname)
		lb.hashMap.addServer(len(lb.replicas)-1, hostname)
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

	lb.hashMap = NewConsistentHashMap(100, len(lb.replicas), 3)
	for i, hostname := range lb.replicas {
		lb.hashMap.addServer(i, hostname)
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

	path := r.URL.Path

	if path != "/hello" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": fmt.Sprintf("<Error> '%s' endpoint does not exist in server replicas", path),
			"status":  "failure",
		})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Introduce randomness within the range of available servers
	randomComponent := strconv.Itoa(rand.Intn(len(lb.replicas)))
	hashKey := fmt.Sprintf("%s:%s:%s", r.RemoteAddr, path, randomComponent)
	serverIndex := lb.hashMap.mapRequest(hashKey)
	selectedServer := lb.replicas[serverIndex]

	log.Printf("Routing request for %s with random component %s to server %s (index %d)", path, randomComponent, selectedServer, serverIndex)

	port, _ := extractServerPort(selectedServer)
	url := fmt.Sprintf("http://localhost:%d%s", port, path)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to route to server: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response from server: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func extractServerPort(hostname string) (int, error) {
	underscoreIndex := strings.LastIndex(hostname, "_")
	if underscoreIndex == -1 {
		return 0, fmt.Errorf("no underscore found in hostname string")
	}

	numberStr := hostname[underscoreIndex+1:]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		return 0, fmt.Errorf("failed to convert substring to integer: %v", err)
	}

	return 8080 + number, nil
}
