package router

import (
	"net/http"
	"pharmacy/adapter/http/middleware"
	"pharmacy/adapter/websocket"
)

// InitWebSocketRouter serves /ws, the live stock-taking updates, to signed-in users. It sits
// outside the main middleware stack because Logging's writer cannot hijack the connection for the
// upgrade; AuthMiddleware passes the writer through unchanged, so it answers a signed-out
// handshake with 401 before the upgrade. Every signed-in user may read a stock taking, so no
// permission is needed here either.
func InitWebSocketRouter(hub *websocket.Hub) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/ws", middleware.AuthMiddleware(http.HandlerFunc(hub.HandleWebSocket)))
	return mux
}
