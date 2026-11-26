package websocket

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gws "github.com/gorilla/websocket"
)

// A socket must not outlive the session that opened it: the cookie that authenticated the
// handshake stops working at the session's end, so the connection closes then too.
func TestLiveCountSocketClosesWhenTheSessionEnds(t *testing.T) {
	saved := sessionEnds
	sessionEnds = func(time.Time) time.Time { return time.Now().Add(100 * time.Millisecond) }
	t.Cleanup(func() { sessionEnds = saved })

	srv := httptest.NewServer(http.HandlerFunc(StartHub().HandleWebSocket))
	t.Cleanup(srv.Close)
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?stockTakingId=1"
	conn, _, err := gws.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, _, err = conn.ReadMessage()
	var netErr net.Error
	if err == nil || (errors.As(err, &netErr) && netErr.Timeout()) {
		t.Fatalf("the server should close the socket at the session's end, got %v", err)
	}
}
