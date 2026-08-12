package ws_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TestHub_UnregisterBeforeRegister_NoZombie_P1_BE_21 — W3-A4-001:
// если removeClient (Unregister) закрыл done до registerClient, клиент
// не должен попасть в clients/userClients (зомби).
//
// Mutation: убрать select{case <-client.done: return} в registerClient →
// ExportHasConnection станет true после «Unregister → Register».
func TestHub_UnregisterBeforeRegister_NoZombie_P1_BE_21(t *testing.T) {
	hub := ws.NewHub()
	client := ws.NewClient("user-race", "session-race")

	// Симулируем порядок Unregister → Register в горутине Run.
	ws.ExportRemoveClient(hub, client)
	select {
	case <-ws.ExportClientDone(client):
	default:
		t.Fatal("expected done closed after Unregister/removeClient")
	}

	ws.ExportRegisterClient(hub, client)

	assert.False(t, ws.ExportHasConnection(hub, client.ConnectionID),
		"zombie: client must not remain in clients after done closed")
	assert.Equal(t, 0, ws.ExportConnectionsForUser(hub, "user-race"))
}

// TestHub_ChannelUnregisterBeforeRegister_NoZombie_P1_BE_21 — тот же инвариант
// через публичные Register/Unregister (канал → Run). Повторяем churn, пока
// select не обработает unregister раньше register хотя бы раз, либо пока
// все пути не очистят клиента.
func TestHub_ChannelUnregisterBeforeRegister_NoZombie_P1_BE_21(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	for i := 0; i < 50; i++ {
		client := ws.NewClient("user-churn", "session-churn")
		hub.Register(client)
		hub.Unregister(client)

		require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
			select {
			case <-ws.ExportClientDone(client):
				return true
			default:
				return false
			}
		}), "done never closed")

		require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
			return !ws.ExportHasConnection(hub, client.ConnectionID) &&
				ws.ExportConnectionsForUser(hub, "user-churn") == 0
		}), "zombie client left in hub maps after Unregister")
	}

	cancel()
	runWG.Wait()
}

// TestHub_LateSubscribeAfterUnregister_NoRoomRevive — W3-A4-002 companion:
// Subscribe после Unregister не возвращает мёртвый Client в rooms.
func TestHub_LateSubscribeAfterUnregister_NoRoomRevive(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		hub.Run(ctx)
	}()

	dead := ws.NewClient("user-late-sub", "session-late")
	ws.ExportRemoveClient(hub, dead) // close done без register
	hub.Subscribe(dead, "room-zombie")

	require.True(t, waitFor(t, 500*time.Millisecond, func() bool {
		// Subscribe обработан Run'ом и отклонён — комнаты пусты.
		return !ws.ExportClientInRoom(dead, "room-zombie")
	}))
	assert.False(t, ws.ExportHasConnection(hub, dead.ConnectionID))

	cancel()
	wg.Wait()
}
