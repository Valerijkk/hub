package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/ws"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHub_MultiTab_ChannelBackedRegisterAndSendToUser проверяет, что когда
// клиенты регистрируются через публичные методы Register/Unregister (которые
// идут через каналы, как в production), SendToUser достигает оба активных
// соединения одного и того же пользователя.
//
// LO-06: гард-тест для HI-07 (multi-tab fan-out). Существующие тесты в
// protocol_test.go обращаются к registerClient напрямую — здесь же мы
// прогоняем поток через Run-горутину, чтобы будущий рефакторинг каналов
// не сломал реальный путь регистрации.
func TestHub_MultiTab_ChannelBackedRegisterAndSendToUser(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	a := ws.NewClient("user-1", "session-A")
	b := ws.NewClient("user-1", "session-B")
	assert.NotEqual(t, a.ConnectionID, b.ConnectionID, "ConnectionID должен быть уникален для каждого таба")

	hub.Register(a)
	hub.Register(b)

	// Дать Run-горутине обработать оба register-сообщения. Каналы буферизованы,
	// поэтому Register не блокирует — нужно явное ожидание.
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-1") == 2
	}) {
		t.Fatalf("hub never observed both connections for user-1: got=%d", ws.ExportConnectionsForUser(hub, "user-1"))
	}

	hub.SendToUser("user-1", ws.ServerMessage{Type: ws.TypePong})

	for _, c := range []*ws.Client{a, b} {
		select {
		case <-ws.ExportClientSendCh(c):
			// ок
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client conn=%s did not receive fan-out message", c.ConnectionID)
		}
	}

	cancel()
	runWG.Wait()
}

// waitFor опрашивает condition через короткие интервалы, пока она не вернёт
// true или не истечёт timeout. Возвращает true если condition выполнилось.
func waitFor(_ *testing.T, timeout time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return condition()
}

// TestHub_SEC05_ConcurrentSendersDuringRemoveDoNotPanic — гарантирует, что
// параллельные writer'ы (BroadcastToRoom, SendToUser), пишущие в ws.ExportClientSendCh(client)
// одновременно с removeClient, не вызывают "send on closed channel" panic.
//
// До фикса SEC-05 receiver (ws.Hub.removeClient) закрывал send-канал, в который
// одновременно писали bridge/handler-горутины — гонка приводила к панике.
// После фикса removeClient закрывает только done, и writer'ы корректно
// видят сигнал через select { case <-done: }.
func TestHub_SEC05_ConcurrentSendersDuringRemoveDoNotPanic(t *testing.T) {
	const senders = 100
	const iters = 50

	for trial := 0; trial < 20; trial++ {
		hub := ws.NewHub()
		ctx, cancel := context.WithCancel(context.Background())

		var runWG sync.WaitGroup
		runWG.Add(1)
		go func() {
			defer runWG.Done()
			hub.Run(ctx)
		}()

		client := ws.NewClient("user-sec05", "session-sec05")
		hub.Register(client)
		hub.Subscribe(client, "room-sec05")

		// Дать Run обработать register+subscribe.
		if !waitFor(t, 500*time.Millisecond, func() bool {
			return ws.ExportConnectionsForUser(hub, "user-sec05") == 1
		}) {
			cancel()
			runWG.Wait()
			t.Fatalf("client did not register in time")
		}

		// Постоянный дренаж send, иначе буфер забьётся и writer'ы пойдут в default.
		drainerDone := make(chan struct{})
		var drainerWG sync.WaitGroup
		drainerWG.Add(1)
		go func() {
			defer drainerWG.Done()
			for {
				select {
				case <-drainerDone:
					return
				case <-ws.ExportClientSendCh(client):
				case <-ws.ExportClientDone(client):
					// После закрытия done продолжаем дренировать остатки,
					// но выходим через drainerDone, когда тест разрешит.
				}
			}
		}()

		var sendersWG sync.WaitGroup
		sendersWG.Add(senders)
		for i := 0; i < senders; i++ {
			go func() {
				defer sendersWG.Done()
				for j := 0; j < iters; j++ {
					hub.BroadcastToRoom("room-sec05", ws.ServerMessage{Type: ws.TypeEvent, Event: "tick"})
					hub.SendToUser("user-sec05", ws.ServerMessage{Type: ws.TypeEvent, Event: "tick"})
				}
			}()
		}

		// Параллельно с активной записью снимаем клиента — это и есть
		// та самая гонка, которая ронялась паникой до SEC-05.
		go hub.Unregister(client)

		sendersWG.Wait()
		close(drainerDone)
		drainerWG.Wait()

		cancel()
		runWG.Wait()

		// done должен быть закрыт по факту Unregister.
		select {
		case <-ws.ExportClientDone(client):
			// ок
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("client.done was not closed after Unregister (trial=%d)", trial)
		}
	}
}

// TestHub_RemoveClient_CleansMembership проверяет сценарий: hub remove client cleans membership.
func TestHub_RemoveClient_CleansMembership(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	client := ws.NewClient("user-rm", "session-rm")
	hub.Register(client)
	hub.Subscribe(client, "room-rm")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-rm") == 1
	}) {
		t.Fatal("client did not register")
	}

	hub.Unregister(client)

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-rm") == 0
	}) {
		t.Fatalf("user still registered: count=%d", ws.ExportConnectionsForUser(hub, "user-rm"))
	}

	select {
	case <-ws.ExportClientDone(client):
	case <-time.After(500 * time.Millisecond):
		t.Fatal("client.done not closed after remove")
	}

	hub.BroadcastToRoom("room-rm", ws.ServerMessage{Type: ws.TypeEvent, Event: "after-remove"})
	select {
	case msg := <-ws.ExportClientSendCh(client):
		t.Fatalf("removed client received broadcast: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	runWG.Wait()
}

// TestHub_BroadcastToRoom_DeliversToSubscribers проверяет сценарий: hub broadcast to room delivers to subscribers.
func TestHub_BroadcastToRoom_DeliversToSubscribers(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	a := ws.NewClient("user-a", "session-a")
	b := ws.NewClient("user-b", "session-b")
	hub.Register(a)
	hub.Register(b)
	hub.Subscribe(a, "room-broadcast")
	hub.Subscribe(b, "room-broadcast")
	hub.Subscribe(b, "room-other")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-a") == 1 && ws.ExportConnectionsForUser(hub, "user-b") == 1
	}) {
		t.Fatal("clients did not register")
	}

	hub.BroadcastToRoom("room-broadcast", ws.ServerMessage{Type: ws.TypeEvent, Event: "ping"})

	for _, c := range []*ws.Client{a, b} {
		select {
		case raw := <-ws.ExportClientSendCh(c):
			var msg ws.ServerMessage
			require.NoError(t, json.Unmarshal(raw, &msg))
			assert.Equal(t, ws.TypeEvent, msg.Type)
			assert.Equal(t, "ping", msg.Event)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client %s did not receive broadcast", c.ConnectionID)
		}
	}

	cancel()
	runWG.Wait()
}

// TestHub_BroadcastToRoom_EmptyRoomNoPanic гарантирует отсутствие паники/гонок при hub broadcast to room пустой room no паника.
func TestHub_BroadcastToRoom_EmptyRoomNoPanic(t *testing.T) {
	hub := ws.NewHub()
	assert.NotPanics(t, func() {
		hub.BroadcastToRoom("nonexistent-room", ws.ServerMessage{Type: ws.TypePong})
	})
}

// TestHub_SendToUser_UnknownUserNoPanic гарантирует отсутствие паники/гонок при hub отправка to user unknown user no паника.
func TestHub_SendToUser_UnknownUserNoPanic(t *testing.T) {
	hub := ws.NewHub()
	assert.NotPanics(t, func() {
		hub.SendToUser("nobody-here", ws.ServerMessage{Type: ws.TypePong})
	})
}

// TestHub_SendToUser_DropsWhenBufferFull проверяет сценарий: hub отправка to user drops when buffer full.
func TestHub_SendToUser_DropsWhenBufferFull(t *testing.T) {
	hub := ws.NewHub()
	client := ws.NewClient("user-drop", "session-drop")
	ws.ExportRegisterClient(hub, client)

	for i := 0; i < cap(ws.ExportClientSendCh(client)); i++ {
		ws.ExportClientSendCh(client) <- []byte("fill")
	}
	queuedBefore := len(ws.ExportClientSendCh(client))

	assert.NotPanics(t, func() {
		hub.SendToUser("user-drop", ws.ServerMessage{Type: ws.TypeEvent, Event: "overflow"})
	})
	assert.Equal(t, queuedBefore, len(ws.ExportClientSendCh(client)), "overflow message must be dropped when buffer is full")
}

// TestHub_BroadcastToRoom_MultiTabSameRoom проверяет сценарий: hub broadcast to room multi tab same room.
func TestHub_BroadcastToRoom_MultiTabSameRoom(t *testing.T) {
	hub, _ := startTestHub(t)

	a := ws.NewClient("user-mt", "tab-a")
	b := ws.NewClient("user-mt", "tab-b")
	hub.Register(a)
	hub.Register(b)
	hub.Subscribe(a, "room-mt")
	hub.Subscribe(b, "room-mt")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-mt") == 2
	}) {
		t.Fatal("both tabs did not register")
	}

	hub.BroadcastToRoom("room-mt", ws.ServerMessage{Type: ws.TypeEvent, Event: "multi"})

	for _, c := range []*ws.Client{a, b} {
		select {
		case raw := <-ws.ExportClientSendCh(c):
			var msg ws.ServerMessage
			require.NoError(t, json.Unmarshal(raw, &msg))
			assert.Equal(t, "multi", msg.Event)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("tab %s did not receive broadcast", c.ConnectionID)
		}
	}
}

// TestHub_Unsubscribe_StopsBroadcast проверяет сценарий: hub unsubscribe stops broadcast.
func TestHub_Unsubscribe_StopsBroadcast(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	client := ws.NewClient("user-unsub", "session-unsub")
	hub.Register(client)
	hub.Subscribe(client, "room-unsub")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-unsub") == 1
	}) {
		t.Fatal("client did not register")
	}

	hub.Unsubscribe(client, "room-unsub")
	time.Sleep(20 * time.Millisecond)

	hub.BroadcastToRoom("room-unsub", ws.ServerMessage{Type: ws.TypeEvent, Event: "after-unsub"})
	select {
	case msg := <-ws.ExportClientSendCh(client):
		t.Fatalf("unsubscribed client received message: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	runWG.Wait()
}
