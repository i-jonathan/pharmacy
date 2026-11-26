package websocket

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

// upgrader leaves CheckOrigin nil, so gorilla's default applies: the Origin's host must equal the
// request's Host. The socket is authenticated by the session cookie, which the browser sends from
// any site, so this is what stops another page the cashier opens from reading live counts
// (cross-site WebSocket hijacking). A handshake without an Origin is not from a browser and still
// needs the session.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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
	if err != nil || stockTakingID <= 0 {
		log.Printf("Invalid stockTakingId parameter: %q", stockTakingIDStr)
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
