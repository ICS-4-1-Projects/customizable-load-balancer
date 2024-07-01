package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/heartbeat", heartbeatHandler)

	serverId := getServerId()
	log.Printf("Server %s is running on port 5000", serverId)

	log.Fatal(http.ListenAndServe(":5000", nil))
}

func getServerId() string {
	serverID := os.Getenv("SERVER_ID")

	if serverID == "" {
		serverID = "unknown"
	}

	return serverID
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	serverID := getServerId()
	response := fmt.Sprintf("Hello from %s", serverID)

	fmt.Fprint(w, response)
}

func heartbeatHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
