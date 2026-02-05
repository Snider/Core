package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	t.Run("creates hub with initialized maps", func(t *testing.T) {
		hub := NewHub()

		require.NotNil(t, hub)
		assert.NotNil(t, hub.clients)
		assert.NotNil(t, hub.broadcast)
		assert.NotNil(t, hub.register)
		assert.NotNil(t, hub.unregister)
		assert.NotNil(t, hub.channels)
	})
}

func TestHub_Run(t *testing.T) {
	t.Run("stops on context cancel", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())

		done := make(chan struct{})
		go func() {
			hub.Run(ctx)
			close(done)
		}()

		cancel()

		select {
		case <-done:
			// Good - hub stopped
		case <-time.After(time.Second):
			t.Fatal("hub should have stopped on context cancel")
		}
	})
}

func TestHub_Broadcast(t *testing.T) {
	t.Run("marshals message with timestamp", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		msg := Message{
			Type: TypeEvent,
			Data: "test data",
		}

		err := hub.Broadcast(msg)
		require.NoError(t, err)
	})

	t.Run("returns error when channel full", func(t *testing.T) {
		hub := NewHub()
		// Fill the broadcast channel
		for i := 0; i < 256; i++ {
			hub.broadcast <- []byte("test")
		}

		err := hub.Broadcast(Message{Type: TypeEvent})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "broadcast channel full")
	})
}

func TestHub_Stats(t *testing.T) {
	t.Run("returns empty stats for new hub", func(t *testing.T) {
		hub := NewHub()

		stats := hub.Stats()

		assert.Equal(t, 0, stats.Clients)
		assert.Equal(t, 0, stats.Channels)
	})

	t.Run("tracks client and channel counts", func(t *testing.T) {
		hub := NewHub()

		// Manually add clients for testing
		hub.mu.Lock()
		client1 := &Client{subscriptions: make(map[string]bool)}
		client2 := &Client{subscriptions: make(map[string]bool)}
		hub.clients[client1] = true
		hub.clients[client2] = true
		hub.channels["test-channel"] = make(map[*Client]bool)
		hub.mu.Unlock()

		stats := hub.Stats()

		assert.Equal(t, 2, stats.Clients)
		assert.Equal(t, 1, stats.Channels)
	})
}

func TestHub_ClientCount(t *testing.T) {
	t.Run("returns zero for empty hub", func(t *testing.T) {
		hub := NewHub()
		assert.Equal(t, 0, hub.ClientCount())
	})

	t.Run("counts connected clients", func(t *testing.T) {
		hub := NewHub()

		hub.mu.Lock()
		hub.clients[&Client{}] = true
		hub.clients[&Client{}] = true
		hub.mu.Unlock()

		assert.Equal(t, 2, hub.ClientCount())
	})
}

func TestHub_ChannelCount(t *testing.T) {
	t.Run("returns zero for empty hub", func(t *testing.T) {
		hub := NewHub()
		assert.Equal(t, 0, hub.ChannelCount())
	})

	t.Run("counts active channels", func(t *testing.T) {
		hub := NewHub()

		hub.mu.Lock()
		hub.channels["channel1"] = make(map[*Client]bool)
		hub.channels["channel2"] = make(map[*Client]bool)
		hub.mu.Unlock()

		assert.Equal(t, 2, hub.ChannelCount())
	})
}

func TestHub_ChannelSubscriberCount(t *testing.T) {
	t.Run("returns zero for non-existent channel", func(t *testing.T) {
		hub := NewHub()
		assert.Equal(t, 0, hub.ChannelSubscriberCount("non-existent"))
	})

	t.Run("counts subscribers in channel", func(t *testing.T) {
		hub := NewHub()

		hub.mu.Lock()
		hub.channels["test-channel"] = make(map[*Client]bool)
		hub.channels["test-channel"][&Client{}] = true
		hub.channels["test-channel"][&Client{}] = true
		hub.mu.Unlock()

		assert.Equal(t, 2, hub.ChannelSubscriberCount("test-channel"))
	})
}

func TestHub_Subscribe(t *testing.T) {
	t.Run("adds client to channel", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		hub.mu.Lock()
		hub.clients[client] = true
		hub.mu.Unlock()

		hub.Subscribe(client, "test-channel")

		assert.Equal(t, 1, hub.ChannelSubscriberCount("test-channel"))
		assert.True(t, client.subscriptions["test-channel"])
	})

	t.Run("creates channel if not exists", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		hub.Subscribe(client, "new-channel")

		hub.mu.RLock()
		_, exists := hub.channels["new-channel"]
		hub.mu.RUnlock()

		assert.True(t, exists)
	})
}

func TestHub_Unsubscribe(t *testing.T) {
	t.Run("removes client from channel", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		hub.Subscribe(client, "test-channel")
		assert.Equal(t, 1, hub.ChannelSubscriberCount("test-channel"))

		hub.Unsubscribe(client, "test-channel")
		assert.Equal(t, 0, hub.ChannelSubscriberCount("test-channel"))
		assert.False(t, client.subscriptions["test-channel"])
	})

	t.Run("cleans up empty channels", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		hub.Subscribe(client, "temp-channel")
		hub.Unsubscribe(client, "temp-channel")

		hub.mu.RLock()
		_, exists := hub.channels["temp-channel"]
		hub.mu.RUnlock()

		assert.False(t, exists, "empty channel should be removed")
	})

	t.Run("handles non-existent channel gracefully", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		// Should not panic
		hub.Unsubscribe(client, "non-existent")
	})
}

func TestHub_SendToChannel(t *testing.T) {
	t.Run("sends to channel subscribers", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		hub.mu.Lock()
		hub.clients[client] = true
		hub.mu.Unlock()
		hub.Subscribe(client, "test-channel")

		err := hub.SendToChannel("test-channel", Message{
			Type: TypeEvent,
			Data: "test",
		})
		require.NoError(t, err)

		select {
		case msg := <-client.send:
			var received Message
			err := json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeEvent, received.Type)
			assert.Equal(t, "test-channel", received.Channel)
		case <-time.After(time.Second):
			t.Fatal("expected message on client send channel")
		}
	})

	t.Run("returns nil for non-existent channel", func(t *testing.T) {
		hub := NewHub()

		err := hub.SendToChannel("non-existent", Message{Type: TypeEvent})
		assert.NoError(t, err, "should not error for non-existent channel")
	})
}

func TestHub_SendProcessOutput(t *testing.T) {
	t.Run("sends output to process channel", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		hub.mu.Lock()
		hub.clients[client] = true
		hub.mu.Unlock()
		hub.Subscribe(client, "process:proc-1")

		err := hub.SendProcessOutput("proc-1", "hello world")
		require.NoError(t, err)

		select {
		case msg := <-client.send:
			var received Message
			err := json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeProcessOutput, received.Type)
			assert.Equal(t, "proc-1", received.ProcessID)
			assert.Equal(t, "hello world", received.Data)
		case <-time.After(time.Second):
			t.Fatal("expected message on client send channel")
		}
	})
}

func TestHub_SendProcessStatus(t *testing.T) {
	t.Run("sends status to process channel", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		hub.mu.Lock()
		hub.clients[client] = true
		hub.mu.Unlock()
		hub.Subscribe(client, "process:proc-1")

		err := hub.SendProcessStatus("proc-1", "exited", 0)
		require.NoError(t, err)

		select {
		case msg := <-client.send:
			var received Message
			err := json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeProcessStatus, received.Type)
			assert.Equal(t, "proc-1", received.ProcessID)

			data, ok := received.Data.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "exited", data["status"])
			assert.Equal(t, float64(0), data["exitCode"])
		case <-time.After(time.Second):
			t.Fatal("expected message on client send channel")
		}
	})
}

func TestHub_SendError(t *testing.T) {
	t.Run("broadcasts error message", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		hub.register <- client
		// Give time for registration
		time.Sleep(10 * time.Millisecond)

		err := hub.SendError("something went wrong")
		require.NoError(t, err)

		select {
		case msg := <-client.send:
			var received Message
			err := json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeError, received.Type)
			assert.Equal(t, "something went wrong", received.Data)
		case <-time.After(time.Second):
			t.Fatal("expected error message on client send channel")
		}
	})
}

func TestHub_SendEvent(t *testing.T) {
	t.Run("broadcasts event message", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		err := hub.SendEvent("user_joined", map[string]string{"user": "alice"})
		require.NoError(t, err)

		select {
		case msg := <-client.send:
			var received Message
			err := json.Unmarshal(msg, &received)
			require.NoError(t, err)
			assert.Equal(t, TypeEvent, received.Type)

			data, ok := received.Data.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "user_joined", data["event"])
		case <-time.After(time.Second):
			t.Fatal("expected event message on client send channel")
		}
	})
}

func TestClient_Subscriptions(t *testing.T) {
	t.Run("returns copy of subscriptions", func(t *testing.T) {
		hub := NewHub()
		client := &Client{
			hub:           hub,
			subscriptions: make(map[string]bool),
		}

		hub.Subscribe(client, "channel1")
		hub.Subscribe(client, "channel2")

		subs := client.Subscriptions()

		assert.Len(t, subs, 2)
		assert.Contains(t, subs, "channel1")
		assert.Contains(t, subs, "channel2")
	})
}

func TestMessage_JSON(t *testing.T) {
	t.Run("marshals correctly", func(t *testing.T) {
		msg := Message{
			Type:      TypeProcessOutput,
			Channel:   "process:1",
			ProcessID: "1",
			Data:      "output line",
			Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(msg)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"type":"process_output"`)
		assert.Contains(t, string(data), `"channel":"process:1"`)
		assert.Contains(t, string(data), `"processId":"1"`)
		assert.Contains(t, string(data), `"data":"output line"`)
	})

	t.Run("unmarshals correctly", func(t *testing.T) {
		jsonStr := `{"type":"subscribe","data":"channel:test"}`

		var msg Message
		err := json.Unmarshal([]byte(jsonStr), &msg)
		require.NoError(t, err)

		assert.Equal(t, TypeSubscribe, msg.Type)
		assert.Equal(t, "channel:test", msg.Data)
	})
}

func TestHub_WebSocketHandler(t *testing.T) {
	t.Run("upgrades connection and registers client", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Give time for registration
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 1, hub.ClientCount())
	})

	t.Run("handles subscribe message", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Send subscribe message
		subscribeMsg := Message{
			Type: TypeSubscribe,
			Data: "test-channel",
		}
		err = conn.WriteJSON(subscribeMsg)
		require.NoError(t, err)

		// Give time for subscription
		time.Sleep(50 * time.Millisecond)

		assert.Equal(t, 1, hub.ChannelSubscriberCount("test-channel"))
	})

	t.Run("handles unsubscribe message", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Subscribe first
		err = conn.WriteJSON(Message{Type: TypeSubscribe, Data: "test-channel"})
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, hub.ChannelSubscriberCount("test-channel"))

		// Unsubscribe
		err = conn.WriteJSON(Message{Type: TypeUnsubscribe, Data: "test-channel"})
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 0, hub.ChannelSubscriberCount("test-channel"))
	})

	t.Run("responds to ping with pong", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Give time for registration
		time.Sleep(50 * time.Millisecond)

		// Send ping
		err = conn.WriteJSON(Message{Type: TypePing})
		require.NoError(t, err)

		// Read pong response
		var response Message
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&response)
		require.NoError(t, err)

		assert.Equal(t, TypePong, response.Type)
	})

	t.Run("broadcasts messages to clients", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		// Give time for registration
		time.Sleep(50 * time.Millisecond)

		// Broadcast a message
		err = hub.Broadcast(Message{
			Type: TypeEvent,
			Data: "broadcast test",
		})
		require.NoError(t, err)

		// Read the broadcast
		var response Message
		conn.SetReadDeadline(time.Now().Add(time.Second))
		err = conn.ReadJSON(&response)
		require.NoError(t, err)

		assert.Equal(t, TypeEvent, response.Type)
		assert.Equal(t, "broadcast test", response.Data)
	})

	t.Run("unregisters client on connection close", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		// Wait for registration
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, hub.ClientCount())

		// Close connection
		conn.Close()

		// Wait for unregistration
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 0, hub.ClientCount())
	})

	t.Run("removes client from channels on disconnect", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		server := httptest.NewServer(hub.Handler())
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		// Subscribe to channel
		err = conn.WriteJSON(Message{Type: TypeSubscribe, Data: "test-channel"})
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, hub.ChannelSubscriberCount("test-channel"))

		// Close connection
		conn.Close()
		time.Sleep(50 * time.Millisecond)

		// Channel should be cleaned up
		assert.Equal(t, 0, hub.ChannelSubscriberCount("test-channel"))
	})
}

func TestHub_Concurrency(t *testing.T) {
	t.Run("handles concurrent subscriptions", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		var wg sync.WaitGroup
		numClients := 100

		for i := 0; i < numClients; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				client := &Client{
					hub:           hub,
					send:          make(chan []byte, 256),
					subscriptions: make(map[string]bool),
				}

				hub.mu.Lock()
				hub.clients[client] = true
				hub.mu.Unlock()

				hub.Subscribe(client, "shared-channel")
				hub.Subscribe(client, "shared-channel") // Double subscribe should be safe
			}(i)
		}

		wg.Wait()

		assert.Equal(t, numClients, hub.ChannelSubscriberCount("shared-channel"))
	})

	t.Run("handles concurrent broadcasts", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		client := &Client{
			hub:           hub,
			send:          make(chan []byte, 1000),
			subscriptions: make(map[string]bool),
		}

		hub.register <- client
		time.Sleep(10 * time.Millisecond)

		var wg sync.WaitGroup
		numBroadcasts := 100

		for i := 0; i < numBroadcasts; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_ = hub.Broadcast(Message{
					Type: TypeEvent,
					Data: id,
				})
			}(i)
		}

		wg.Wait()

		// Give time for broadcasts to be delivered
		time.Sleep(100 * time.Millisecond)

		// Count received messages
		received := 0
		timeout := time.After(100 * time.Millisecond)
	loop:
		for {
			select {
			case <-client.send:
				received++
			case <-timeout:
				break loop
			}
		}

		// All or most broadcasts should be received
		assert.GreaterOrEqual(t, received, numBroadcasts-10, "should receive most broadcasts")
	})
}

func TestHub_HandleWebSocket(t *testing.T) {
	t.Run("alias works same as Handler", func(t *testing.T) {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go hub.Run(ctx)

		// Test with HandleWebSocket directly
		server := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, 1, hub.ClientCount())
	})
}

func TestMustMarshal(t *testing.T) {
	t.Run("marshals valid data", func(t *testing.T) {
		data := mustMarshal(Message{Type: TypePong})
		assert.Contains(t, string(data), "pong")
	})

	t.Run("handles unmarshalable data without panic", func(t *testing.T) {
		// Create a channel which cannot be marshaled
		// This should not panic, even if it returns nil
		ch := make(chan int)
		assert.NotPanics(t, func() {
			_ = mustMarshal(ch)
		})
	})
}
