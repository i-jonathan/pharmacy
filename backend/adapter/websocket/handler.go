package websocket

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from same origin and localhost (dev and prod)
		origin := r.Header.Get("Origin")
		host := r.Host

		// Allow empty origin (direct connections)
		if origin == "" {
			return true
		}

		// Allow same origin
		if origin == host || origin == "http://"+host || origin == "https://"+host {
			return true
		}

		// Allow all localhost connections (dev and prod)
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			return true
		}

		// Allow connections from localhost origins (different ports)
		if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "https://localhost") ||
			strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://127.0.0.1") {
			return true
		}

		return false
	},
}

// HandleWebSocket handles WebSocket connections for stock taking
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get stock taking ID from query parameter
	stockTakingIDStr := r.URL.Query().Get("stockTakingId")
	if stockTakingIDStr == "" {
		log.Printf("WebSocket connection missing stockTakingId parameter")
		http.Error(w, "Missing stockTakingId parameter", http.StatusBadRequest)
		return
	}

	stockTakingID, err := strconv.Atoi(stockTakingIDStr)
	if err != nil {
		log.Printf("Invalid stockTakingId parameter: %v", err)
		http.Error(w, "Invalid stockTakingId parameter", http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client and register with hub
	client := &Client{
		hub:           h,
		conn:          conn,
		send:          make(chan Message, 256),
		stockTakingID: stockTakingID,
	}

	client.hub.register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket connection established for stock taking %d", stockTakingID)
}

// StartHub initializes and starts a WebSocket hub
func StartHub() *Hub {
	hub := NewHub()
	go hub.Run()
	return hub
}
