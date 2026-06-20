package calls

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
	"github.com/google/uuid"
)

type memChatStore struct {
	mu   sync.Mutex
	msgs []CallChatMessage
}

func newMemChatRepository() *ChatRepository {
	return &ChatRepository{store: &memChatStore{}}
}

func (s *memChatStore) ensureIndexes(_ context.Context) error { return nil }

func (s *memChatStore) insert(_ context.Context, msg *CallChatMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgs = append(s.msgs, *msg)
	return nil
}

func (s *memChatStore) listByCallID(_ context.Context, callID string, limit int) ([]CallChatMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var matched []CallChatMessage
	for _, m := range s.msgs {
		if m.CallID == callID {
			matched = append(matched, m)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].CreatedAt.After(matched[j].CreatedAt)
	})
	if len(matched) > limit {
		matched = matched[:limit]
	}
	return matched, nil
}

func (s *memChatStore) deleteByCallID(_ context.Context, callID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.msgs[:0]
	for _, m := range s.msgs {
		if m.CallID != callID {
			kept = append(kept, m)
		}
	}
	s.msgs = kept
	return nil
}

func TestChatRepository_Mem_SendMessage(t *testing.T) {
	repo := newMemChatRepository()
	ctx := context.Background()

	msg, err := repo.SendMessage(ctx, "call-123", "user-1", "User One", "Привет!")
	require.NoError(t, err)
	require.NotNil(t, msg)
	assert.False(t, msg.ID.IsZero())
	assert.Equal(t, "Привет!", msg.Text)
}

func TestChatRepository_Mem_SendMessage_TextTooLong(t *testing.T) {
	repo := newMemChatRepository()
	longText := make([]byte, 2001)
	for i := range longText {
		longText[i] = 'a'
	}
	_, err := repo.SendMessage(context.Background(), "call-123", "user-1", "User", string(longText))
	require.Error(t, err)
}

func TestChatRepository_Mem_ListMessages(t *testing.T) {
	repo := newMemChatRepository()
	ctx := context.Background()
	base := time.Now().UTC().Add(-time.Minute)
	for i, text := range []string{"Первое", "Второе", "Третье"} {
		err := repo.store.insert(ctx, &CallChatMessage{
			ID: bson.NewObjectID(), CallID: "call-list", UserID: "u1", UserName: "U",
			Text: text, CreatedAt: base.Add(time.Duration(i) * time.Second),
		})
		require.NoError(t, err)
	}

	msgs, err := repo.ListMessages(ctx, "call-list", 10)
	require.NoError(t, err)
	require.Len(t, msgs, 3)
	assert.Equal(t, "Первое", msgs[0].Text)
	assert.Equal(t, "Третье", msgs[2].Text)

	msgs, err = repo.ListMessages(ctx, "call-list", 2)
	require.NoError(t, err)
	assert.Len(t, msgs, 2)
}

func TestChatRepository_Mem_DeleteCallChat(t *testing.T) {
	repo := newMemChatRepository()
	ctx := context.Background()
	for _, pair := range []struct{ callID, text string }{
		{"call-del", "1"}, {"call-del", "2"}, {"call-other", "x"},
	} {
		_, err := repo.SendMessage(ctx, pair.callID, "u1", "U", pair.text)
		require.NoError(t, err)
	}

	require.NoError(t, repo.DeleteCallChat(ctx, "call-del"))
	msgs, err := repo.ListMessages(ctx, "call-del", 10)
	require.NoError(t, err)
	assert.Empty(t, msgs)

	msgs, err = repo.ListMessages(ctx, "call-other", 10)
	require.NoError(t, err)
	assert.Len(t, msgs, 1)
}

func TestChatRepository_Mem_EnsureIndexes(t *testing.T) {
	repo := newMemChatRepository()
	require.NoError(t, repo.EnsureIndexes(context.Background()))
}

func TestChatRepository_Mem_ListMessages_DefaultLimit(t *testing.T) {
	repo := newMemChatRepository()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_, err := repo.SendMessage(ctx, "call-lim", "u1", "U", "msg")
		require.NoError(t, err)
	}
	msgs, err := repo.ListMessages(ctx, "call-lim", 0)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)

	msgs, err = repo.ListMessages(ctx, "call-lim", 250)
	require.NoError(t, err)
	assert.Len(t, msgs, 3)
}

type failInsertChatStore struct{ memChatStore }

func (s *failInsertChatStore) insert(context.Context, *CallChatMessage) error {
	return assert.AnError
}

func TestChatRepository_SendMessage_InsertError(t *testing.T) {
	repo := &ChatRepository{store: &failInsertChatStore{}}
	_, err := repo.SendMessage(context.Background(), "c1", "u1", "U", "hi")
	require.Error(t, err)
}

func TestCallChatService_SendMessage_NotConfigured(t *testing.T) {
	f := newCallsTestFixture(t)
	callEnt := f.createChannelCall(t, f.owner)
	_, err := f.svc.chat.SendMessage(context.Background(), callEnt.ID.String(), f.owner.ID.String(), "Owner", "x", false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не настроен")
}

type failDeleteChatStore struct{ memChatStore }

func (s *failDeleteChatStore) deleteByCallID(context.Context, string) error {
	return assert.AnError
}

func TestCallChatService_DeleteCallChat_LogsError(t *testing.T) {
	f := newCallsTestFixture(t)
	f.svc.chat = NewCallChatService(f.svc.repo, &ChatRepository{store: &failDeleteChatStore{}}, f.svc.pub)
	f.svc.chat.DeleteCallChat(context.Background(), uuid.New().String())
}
