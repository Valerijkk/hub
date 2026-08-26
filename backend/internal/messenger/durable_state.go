package messenger

import "time"

// DurableMongoIndexSpec is the migration-independent description of a MongoDB
// index required by the Phase 151 durable Messenger model.
type DurableMongoIndexSpec struct {
	Collection string
	Keys       []string
	Unique     bool
	Purpose    string
}

// MessagePosition is the total-order cursor used by read state. MessageID is
// the deterministic tie-breaker for messages sharing the same UTC timestamp.
type MessagePosition struct {
	CreatedAt time.Time
	MessageID string
}

type ThreadFollowMode string

const (
	ThreadFollowModeAuto           ThreadFollowMode = "auto"
	ThreadFollowModeManualFollow   ThreadFollowMode = "manual_follow"
	ThreadFollowModeManualUnfollow ThreadFollowMode = "manual_unfollow"
)

type ThreadFollowState struct {
	Mode   ThreadFollowMode
	Cursor MessagePosition
}

type ThreadFollowTransition struct {
	Mode     ThreadFollowMode
	Position MessagePosition
}

type PostCreateEffectsState string

const (
	PostCreateEffectsStatePending  PostCreateEffectsState = "pending"
	PostCreateEffectsStateComplete PostCreateEffectsState = "complete"
)

func ExportPhase151DurableMongoIndexSpecs() []DurableMongoIndexSpec {
	specs := []DurableMongoIndexSpec{
		{Collection: "read_states", Keys: []string{"tenant_id", "user_id", "category", "target_kind", "target_id"}, Unique: true, Purpose: "cursor"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "channel_id", "root_message_id"}, Unique: true, Purpose: "identity"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "channel_id"}, Purpose: "cleanup"},
		{Collection: "thread_follows", Keys: []string{"tenant_id", "user_id", "mode", "channel_id"}, Purpose: "user-list"},
		{Collection: "messages", Keys: []string{"channel_id", "created_at", "_id"}, Purpose: "timeline-head"},
		{Collection: "messages", Keys: []string{"channel_id", "thread_id", "created_at", "_id"}, Purpose: "thread"},
		{Collection: "messages", Keys: []string{"channel_id", "resolved_mention_user_ids", "created_at", "_id"}, Purpose: "mentions"},
		{Collection: "messages", Keys: []string{"channel_id", "post_create_effects_state", "created_at", "_id"}, Purpose: "pending-recovery"},
	}
	return specs
}

// ExportPhase151DecideReadStateAdvance applies a forward-only compare-and-set
// decision over the canonical (created_at UTC, message_id) tuple.
func ExportPhase151DecideReadStateAdvance(
	current *MessagePosition,
	candidate MessagePosition,
) (created bool, advanced bool, next MessagePosition) {
	if current == nil {
		return true, false, candidate
	}
	if messagePositionAfter(candidate, *current) {
		return false, true, candidate
	}
	return false, false, *current
}

// ExportPhase151DecideThreadFollowTransition preserves explicit user intent:
// automatic follow can create an initial record but never changes an existing
// auto or manual decision.
func ExportPhase151DecideThreadFollowTransition(
	current *ThreadFollowState,
	transition ThreadFollowTransition,
) (changed bool, next ThreadFollowState) {
	if !isKnownThreadFollowMode(transition.Mode) {
		if current == nil {
			return false, ThreadFollowState{}
		}
		return false, *current
	}

	if current == nil {
		return true, ThreadFollowState{Mode: transition.Mode, Cursor: transition.Position}
	}
	if transition.Mode == ThreadFollowModeAuto {
		return false, *current
	}
	if current.Mode == transition.Mode && messagePositionEqual(current.Cursor, transition.Position) {
		return false, *current
	}
	return true, ThreadFollowState{Mode: transition.Mode, Cursor: transition.Position}
}

func ExportPhase151ShouldClaimPostCreateEffects(result CreateMessageResult) bool {
	return result.Created && result.PostCreateEffectsState == PostCreateEffectsStatePending
}

func ExportPhase151DurableMessageBSONFields() []string {
	return []string{
		"post_create_effects_state",
		"post_create_effects_completed_at",
		"also_send_to_channel",
		"resolved_mention_user_ids",
		"client_nonce_fingerprint",
		"client_nonce_fingerprint_version",
	}
}

func messagePositionAfter(candidate, current MessagePosition) bool {
	candidateTime := candidate.CreatedAt.UTC()
	currentTime := current.CreatedAt.UTC()
	if candidateTime.After(currentTime) {
		return true
	}
	if candidateTime.Before(currentTime) {
		return false
	}
	return candidate.MessageID > current.MessageID
}

func messagePositionEqual(left, right MessagePosition) bool {
	return left.CreatedAt.UTC().Equal(right.CreatedAt.UTC()) && left.MessageID == right.MessageID
}

func isKnownThreadFollowMode(mode ThreadFollowMode) bool {
	switch mode {
	case ThreadFollowModeAuto, ThreadFollowModeManualFollow, ThreadFollowModeManualUnfollow:
		return true
	default:
		return false
	}
}

func newCreateMessageResult(message *MessageDocument, created bool) *CreateMessageResult {
	state := PostCreateEffectsStateComplete
	if message != nil && message.PostCreateEffectsState != "" {
		state = message.PostCreateEffectsState
	}
	return &CreateMessageResult{
		Message: message, Created: created, PostCreateEffectsState: state,
	}
}