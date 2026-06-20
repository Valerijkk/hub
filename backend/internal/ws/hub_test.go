package ws

import (
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
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	a := NewClient("user-1", "session-A")
	b := NewClient("user-1", "session-B")
	assert.NotEqual(t, a.ConnectionID, b.ConnectionID, "ConnectionID должен быть уникален для каждого таба")

	hub.Register(a)
	hub.Register(b)

	// Дать Run-горутине обработать оба register-сообщения. Каналы буферизованы,
	// поэтому Register не блокирует — нужно явное ожидание.
	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-1") == 2
	}) {
		t.Fatalf("hub never observed both connections for user-1: got=%d", hub.connectionsForUser("user-1"))
	}

	hub.SendToUser("user-1", ServerMessage{Type: TypePong})

	for _, c := range []*Client{a, b} {
		select {
		case <-c.send:
			// ok
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client conn=%s did not receive fan-out message", c.ConnectionID)
		}
	}

	cancel()
	runWG.Wait()
}

// connectionsForUser — тестовый хелпер, считает число активных соединений
// зарегистрированного пользователя. Реализован поверх внутренних карт Hub.
func (h *Hub) connectionsForUser(userID string) int {
	val, ok := h.userClients.Load(userID)
	if !ok {
		return 0
	}
	count := 0
	val.(*sync.Map).Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
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
// параллельные writer'ы (BroadcastToRoom, SendToUser), пишущие в client.send
// одновременно с removeClient, не вызывают "send on closed channel" panic.
//
// До фикса SEC-05 receiver (Hub.removeClient) закрывал send-канал, в который
// одновременно писали bridge/handler-горутины — гонка приводила к панике.
// После фикса removeClient закрывает только done, и writer'ы корректно
// видят сигнал через select { case <-done: }.
func TestHub_SEC05_ConcurrentSendersDuringRemoveDoNotPanic(t *testing.T) {
	const senders = 100
	const iters = 50

	for trial := 0; trial < 20; trial++ {
		hub := NewHub()
		ctx, cancel := context.WithCancel(context.Background())

		var runWG sync.WaitGroup
		runWG.Add(1)
		go func() {
			defer runWG.Done()
			hub.Run(ctx)
		}()

		client := NewClient("user-sec05", "session-sec05")
		hub.Register(client)
		hub.Subscribe(client, "room-sec05")

		// Дать Run обработать register+subscribe.
		if !waitFor(t, 500*time.Millisecond, func() bool {
			return hub.connectionsForUser("user-sec05") == 1
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
				case <-client.send:
				case <-client.done:
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
					hub.BroadcastToRoom("room-sec05", ServerMessage{Type: TypeEvent, Event: "tick"})
					hub.SendToUser("user-sec05", ServerMessage{Type: TypeEvent, Event: "tick"})
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
		case <-client.done:
			// ok
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("client.done was not closed after Unregister (trial=%d)", trial)
		}
	}
}

func TestHub_RemoveClient_CleansMembership(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	client := NewClient("user-rm", "session-rm")
	hub.Register(client)
	hub.Subscribe(client, "room-rm")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-rm") == 1
	}) {
		t.Fatal("client did not register")
	}

	hub.Unregister(client)

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-rm") == 0
	}) {
		t.Fatalf("user still registered: count=%d", hub.connectionsForUser("user-rm"))
	}

	select {
	case <-client.done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("client.done not closed after remove")
	}

	hub.BroadcastToRoom("room-rm", ServerMessage{Type: TypeEvent, Event: "after-remove"})
	select {
	case msg := <-client.send:
		t.Fatalf("removed client received broadcast: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	runWG.Wait()
}

func TestHub_BroadcastToRoom_DeliversToSubscribers(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	a := NewClient("user-a", "session-a")
	b := NewClient("user-b", "session-b")
	hub.Register(a)
	hub.Register(b)
	hub.Subscribe(a, "room-broadcast")
	hub.Subscribe(b, "room-broadcast")
	hub.Subscribe(b, "room-other")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-a") == 1 && hub.connectionsForUser("user-b") == 1
	}) {
		t.Fatal("clients did not register")
	}

	hub.BroadcastToRoom("room-broadcast", ServerMessage{Type: TypeEvent, Event: "ping"})

	for _, c := range []*Client{a, b} {
		select {
		case raw := <-c.send:
			var msg ServerMessage
			require.NoError(t, json.Unmarshal(raw, &msg))
			assert.Equal(t, TypeEvent, msg.Type)
			assert.Equal(t, "ping", msg.Event)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("client %s did not receive broadcast", c.ConnectionID)
		}
	}

	cancel()
	runWG.Wait()
}

func TestHub_BroadcastToRoom_EmptyRoomNoPanic(t *testing.T) {
	hub := NewHub()
	assert.NotPanics(t, func() {
		hub.BroadcastToRoom("nonexistent-room", ServerMessage{Type: TypePong})
	})
}

func TestHub_SendToUser_UnknownUserNoPanic(t *testing.T) {
	hub := NewHub()
	assert.NotPanics(t, func() {
		hub.SendToUser("nobody-here", ServerMessage{Type: TypePong})
	})
}

func TestHub_SendToUser_DropsWhenBufferFull(t *testing.T) {
	hub := NewHub()
	client := NewClient("user-drop", "session-drop")
	hub.registerClient(client)

	for i := 0; i < cap(client.send); i++ {
		client.send <- []byte("fill")
	}
	queuedBefore := len(client.send)

	assert.NotPanics(t, func() {
		hub.SendToUser("user-drop", ServerMessage{Type: TypeEvent, Event: "overflow"})
	})
	assert.Equal(t, queuedBefore, len(client.send), "overflow message must be dropped when buffer is full")
}

func TestHub_BroadcastToRoom_MultiTabSameRoom(t *testing.T) {
	hub, _ := startTestHub(t)

	a := NewClient("user-mt", "tab-a")
	b := NewClient("user-mt", "tab-b")
	hub.Register(a)
	hub.Register(b)
	hub.Subscribe(a, "room-mt")
	hub.Subscribe(b, "room-mt")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-mt") == 2
	}) {
		t.Fatal("both tabs did not register")
	}

	hub.BroadcastToRoom("room-mt", ServerMessage{Type: TypeEvent, Event: "multi"})

	for _, c := range []*Client{a, b} {
		select {
		case raw := <-c.send:
			var msg ServerMessage
			require.NoError(t, json.Unmarshal(raw, &msg))
			assert.Equal(t, "multi", msg.Event)
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("tab %s did not receive broadcast", c.ConnectionID)
		}
	}
}

func TestHub_Unsubscribe_StopsBroadcast(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	client := NewClient("user-unsub", "session-unsub")
	hub.Register(client)
	hub.Subscribe(client, "room-unsub")

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return hub.connectionsForUser("user-unsub") == 1
	}) {
		t.Fatal("client did not register")
	}

	hub.Unsubscribe(client, "room-unsub")
	time.Sleep(20 * time.Millisecond)

	hub.BroadcastToRoom("room-unsub", ServerMessage{Type: TypeEvent, Event: "after-unsub"})
	select {
	case msg := <-client.send:
		t.Fatalf("unsubscribed client received message: %q", msg)
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	runWG.Wait()
}
