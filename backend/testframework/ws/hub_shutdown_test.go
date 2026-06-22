package ws_test

// WS-01 / FA-01 P3: graceful Hub shutdown — drain clients, non-blocking ops after stop.
import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/valerii/hub/backend/internal/ws"
)

// TestHub_Shutdown_UnregistersAllClients проверяет, что Run при ctx.Done()
// снимает всех клиентов и закрывает client.done.
func TestHub_Shutdown_UnregistersAllClients(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	const n = 5
	clients := make([]*ws.Client, n)
	for i := range clients {
		clients[i] = ws.NewClient("user-shutdown", fmt.Sprintf("session-%d", i))
		hub.Register(clients[i])
		hub.Subscribe(clients[i], "room-shutdown")
	}

	if !waitFor(t, 500*time.Millisecond, func() bool {
		return ws.ExportConnectionsForUser(hub, "user-shutdown") == n
	}) {
		t.Fatalf("clients did not register: got=%d", ws.ExportConnectionsForUser(hub, "user-shutdown"))
	}

	cancel()
	runWG.Wait()

	for i, c := range clients {
		select {
		case <-ws.ExportClientDone(c):
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("client %d done not closed after hub shutdown", i)
		}
	}

	assert.Equal(t, 0, ws.ExportConnectionsForUser(hub, "user-shutdown"))
}

// TestHub_Shutdown_RegisterSubscribeDoNotBlock гарантирует, что Register/Subscribe/
// Unregister/Unsubscribe не блокируются после остановки Run.
func TestHub_Shutdown_RegisterSubscribeDoNotBlock(t *testing.T) {
	hub := ws.NewHub()
	ctx, cancel := context.WithCancel(context.Background())

	var runWG sync.WaitGroup
	runWG.Add(1)
	go func() {
		defer runWG.Done()
		hub.Run(ctx)
	}()

	cancel()
	runWG.Wait()

	client := ws.NewClient("late-user", "late-session")

	type op struct {
		name string
		fn   func()
	}
	ops := []op{
		{"Register", func() { hub.Register(client) }},
		{"Subscribe", func() { hub.Subscribe(client, "room-late") }},
		{"Unsubscribe", func() { hub.Unsubscribe(client, "room-late") }},
		{"Unregister", func() { hub.Unregister(client) }},
	}

	for _, op := range ops {
		op := op
		t.Run(op.name, func(t *testing.T) {
			done := make(chan struct{})
			go func() {
				op.fn()
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(500 * time.Millisecond):
				t.Fatalf("%s blocked after hub shutdown", op.name)
			}
		})
	}

	select {
	case <-ws.ExportClientDone(client):
	case <-time.After(500 * time.Millisecond):
		t.Fatal("late Register did not signal client.done")
	}
}

// TestHub_Shutdown_ConcurrentOpsNoPanic — гонка cancel + Register/Subscribe/Unregister
// под -race не должна паниковать или зависать.
func TestHub_Shutdown_ConcurrentOpsNoPanic(t *testing.T) {
	const trials = 10
	const workers = 20

	for trial := 0; trial < trials; trial++ {
		hub := ws.NewHub()
		ctx, cancel := context.WithCancel(context.Background())

		var runWG sync.WaitGroup
		runWG.Add(1)
		go func() {
			defer runWG.Done()
			hub.Run(ctx)
		}()

		// Держим хотя бы одного зарегистрированного клиента до cancel.
		seed := ws.NewClient("seed-user", "seed-session")
		hub.Register(seed)
		hub.Subscribe(seed, "room-seed")
		if !waitFor(t, 500*time.Millisecond, func() bool {
			return ws.ExportConnectionsForUser(hub, "seed-user") == 1
		}) {
			cancel()
			runWG.Wait()
			t.Fatalf("seed client did not register (trial=%d)", trial)
		}

		var workersWG sync.WaitGroup
		workersWG.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer workersWG.Done()
				c := ws.NewClient("race-user", "race-session")
				hub.Register(c)
				hub.Subscribe(c, "race-room")
				hub.Unsubscribe(c, "race-room")
				hub.Unregister(c)
			}()
		}

		cancel()
		workersWG.Wait()
		runWG.Wait()

		select {
		case <-ws.ExportClientDone(seed):
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("seed client.done not closed after shutdown (trial=%d)", trial)
		}
	}
}
