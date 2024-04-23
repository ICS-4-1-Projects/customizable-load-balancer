package main

import (
	"fmt"
	"math"
)

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
	return int(math.Pow(float64(serverID), 2)) + int(math.Pow(float64(virtualID), 2)) + (2 * virtualID) + 25
}

func (c *ConsistentHashMap) requestHash(requestID int) int {
	return ((requestID * requestID) + (2 * requestID) + 17) % c.numSlots
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
	hashMap := NewConsistentHashMap(512, 5, int(math.Log2(512)))

	for requestID := 0; requestID < 100; requestID++ {
		server := hashMap.mapRequest(requestID)
		fmt.Printf("Request %d is handled by server container %d\n", requestID, server)
	}
}