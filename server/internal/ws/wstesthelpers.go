package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"atlas.sageyutta.com/internal/wsmessages"
	"github.com/gorilla/websocket"
)

func recvMsg(t testing.TB, c *Client) wsmessages.WSMessageOut {
	t.Helper()
	select {
	case msg := <-c.send:
		return msg
	case <-time.After(time.Second):
		t.Fatalf("client %s: expected a message, timed out", c.id)
		return wsmessages.WSMessageOut{}
	}
}

func expectClosed(t testing.TB, c *Client) {
	t.Helper()
	select {
	case msg, ok := <-c.send:
		if ok {
			t.Fatalf("client %s: expected send channel closed, got message %v", c.id, msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatalf("client %s: send channel was not closed", c.id)
	}
}

func expectNoMsg(t testing.TB, c *Client) {
	t.Helper()
	select {
	case msg, ok := <-c.send:
		if ok {
			t.Fatalf("client %s: expected no message, got %v", c.id, msg.Type)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

var testUpgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

func newWebsocketPair(t testing.TB) (serverConn, clientConn *websocket.Conn) {
	t.Helper()

	serverConnCh := make(chan *websocket.Conn, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := testUpgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("websocket upgrade failed: %v", err)
			return
		}
		serverConnCh <- conn
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial %s: %v", wsURL, err)
	}
	t.Cleanup(func() { clientConn.Close() })

	select {
	case serverConn = <-serverConnCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server-side websocket connection")
	}
	t.Cleanup(func() { serverConn.Close() })

	return serverConn, clientConn
}
