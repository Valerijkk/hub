package ws

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
