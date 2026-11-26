package router

import (
	"net/http"
	"net/http/httptest"
	"pharmacy/adapter/websocket"
	"pharmacy/config"
	"pharmacy/internal/constant"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
)

// signedInCookie returns the session cookie the login handler would set for a user.
func signedInCookie(t *testing.T) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	session, _ := config.NewSessionStore().New(req, "session")
	session.Values[constant.UserSessionKey] = 7
	if err := session.Save(req, rec); err != nil {
		t.Fatal(err)
	}
	c := rec.Result().Cookies()[0]
	return c.Name + "=" + c.Value
}

// dialLiveCount opens the stock-taking socket the way the browser does, returning the
// connection (nil when refused) and the handshake status.
func dialLiveCount(
	t *testing.T,
	srv *httptest.Server,
	query, cookie, origin string,
) (*gws.Conn, int) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?" + query
	header := http.Header{}
	if cookie != "" {
		header.Set("Cookie", cookie)
	}
	if origin != "" {
		header.Set("Origin", origin)
	}
	conn, resp, err := gws.DefaultDialer.Dial(url, header)
	if err == nil {
		t.Cleanup(func() { conn.Close() })
		return conn, resp.StatusCode
	}
	if resp == nil {
		t.Fatalf("dial failed without a response: %v", err)
	}
	return nil, resp.StatusCode
}

func liveCountServer(t *testing.T) (*httptest.Server, *websocket.Hub) {
	t.Helper()
	hub := websocket.StartHub()
	srv := httptest.NewServer(InitWebSocketRouter(hub))
	t.Cleanup(srv.Close)
	return srv, hub
}

func TestLiveCountSocketNeedsASession(t *testing.T) {
	srv, _ := liveCountServer(t)
	conn, status := dialLiveCount(t, srv, "stockTakingId=1", "", srv.URL)
	if conn != nil || status != http.StatusUnauthorized {
		t.Fatalf("a signed-out handshake should get 401 and no socket, got %d", status)
	}
}

func TestLiveCountSocketDeliversUpdatesToASignedInUser(t *testing.T) {
	srv, hub := liveCountServer(t)
	conn, status := dialLiveCount(t, srv, "stockTakingId=4", signedInCookie(t), srv.URL)
	if conn == nil {
		t.Fatalf("a signed-in same-origin handshake should upgrade, got %d", status)
	}
	// The hub registers the client asynchronously; keep broadcasting until one arrives.
	got := make(chan websocket.Message, 1)
	go func() {
		var m websocket.Message
		if conn.ReadJSON(&m) == nil {
			got <- m
		}
	}()
	deadline := time.After(2 * time.Second)
	for {
		hub.BroadcastStockTakingComplete(4)
		select {
		case m := <-got:
			if m.Type != websocket.StockTakingComplete || m.StockTakingID != 4 {
				t.Fatalf("unexpected message %+v", m)
			}
			return
		case <-deadline:
			t.Fatal("no message arrived")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

// Once the socket trusts the session cookie, another site the cashier visits must not be able to
// open it with that cookie, even when the shop's server is on localhost.
func TestLiveCountSocketRefusesAnotherOrigin(t *testing.T) {
	srv, _ := liveCountServer(t)
	for _, origin := range []string{"https://evil.example", "http://localhost:5173"} {
		conn, status := dialLiveCount(t, srv, "stockTakingId=1", signedInCookie(t), origin)
		if conn != nil || status != http.StatusForbidden {
			t.Fatalf("origin %s should be refused with 403, got %d", origin, status)
		}
	}
}

func TestLiveCountSocketNeedsAStockTakingID(t *testing.T) {
	srv, _ := liveCountServer(t)
	for _, q := range []string{"", "stockTakingId=abc", "stockTakingId=0", "stockTakingId=-3"} {
		conn, status := dialLiveCount(t, srv, q, signedInCookie(t), srv.URL)
		if conn != nil || status != http.StatusBadRequest {
			t.Fatalf("query %q should be refused with 400, got %d", q, status)
		}
	}
}
