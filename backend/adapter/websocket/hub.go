package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message types
const (
	StockItemUpdate     = "stock_item_update"
	StockTakingComplete = "stock_taking_complete"
)

// Message represents a WebSocket message
type Message struct {
	StockTakingID int    `json:"stockTakingId,omitempty"`
	Type          string `json:"type"`
	Data          any    `json:"data"`
}

// Client represents a WebSocket client
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan Message
	stockTakingID int
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected for stock taking %d", client.stockTakingID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client disconnected for stock taking %d", client.stockTakingID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				// Send message only to clients interested in this stock taking
				if message.StockTakingID == 0 || client.stockTakingID == message.StockTakingID {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastStockItemUpdate broadcasts a stock item update to relevant clients
func (h *Hub) BroadcastStockItemUpdate(stockTakingID int, item any) {
	message := Message{
		Type:          StockItemUpdate,
		Data:          item,
		StockTakingID: stockTakingID,
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// BroadcastStockTakingComplete broadcasts stock taking completion
func (h *Hub) BroadcastStockTakingComplete(stockTakingID int) {
	message := Message{
		Type:          StockTakingComplete,
		Data:          map[string]any{"stockTakingId": stockTakingID},
		StockTakingID: stockTakingID,
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// WritePump writes messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		if err := c.conn.WriteJSON(message); err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			return
		}
	}

	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Set read deadline and pong handler
	c.conn.SetReadLimit(512)

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}
