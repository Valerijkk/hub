package messenger_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/messenger"
)

func TestPhase151RecipientChannelResponseCopiesFreshAccessCapabilities(t *testing.T) {
	t.Parallel()

	access := phase151RecipientChannelAccess(true)
	response, err := messenger.ExportToRecipientChannelResponse(access)
	if err != nil {
		t.Fatalf("ExportToRecipientChannelResponse() error = %v", err)
	}

	capabilities := phase151CapabilitiesJSON(t, response)
	for key, want := range phase151CapabilityValues(access) {
		got, ok := capabilities[key]
		if !ok {
			t.Fatalf("capabilities.%s is absent; every authorization boolean must be transported", key)
		}
		if got != want {
			t.Fatalf("capabilities.%s = %t, want fresh ChannelAccess value %t", key, got, want)
		}
	}
}

func TestPhase151RecipientChannelResponseDoesNotElevateDeniedCapabilitiesOrExposeAuthorityInternals(t *testing.T) {
	t.Parallel()

	access := phase151RecipientChannelAccess(false)
	access.Role = "owner"
	response, err := messenger.ExportToRecipientChannelResponse(access)
	if err != nil {
		t.Fatalf("ExportToRecipientChannelResponse() error = %v", err)
	}

	capabilities := phase151CapabilitiesJSON(t, response)
	for key, want := range phase151CapabilityValues(access) {
		if want {
			t.Fatalf("test setup must deny %s", key)
		}
		if got, ok := capabilities[key]; !ok || got {
			t.Fatalf("capabilities.%s = %t (present=%t), want false from fresh access despite owner role", key, got, ok)
		}
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal recipient response: %v", err)
	}
	var payload struct {
		Capabilities map[string]json.RawMessage `json:"capabilities"`
	}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("unmarshal recipient response: %v", err)
	}
	for _, forbidden := range []string{"id", "role", "channel_id", "user_id"} {
		if _, ok := payload.Capabilities[forbidden]; ok {
			t.Fatalf("capabilities must not disclose %q", forbidden)
		}
	}
}

func TestPhase151RecipientChannelResponseFailsClosedWithoutChannel(t *testing.T) {
	t.Parallel()

	response, err := messenger.ExportToRecipientChannelResponse(messenger.ChannelAccess{})
	if err == nil {
		t.Fatal("ExportToRecipientChannelResponse() error = nil, want fail-closed error for absent access.Channel")
	}
	if response != nil {
		t.Fatalf("ExportToRecipientChannelResponse() response = %#v, want nil on absent access.Channel", response)
	}
}

func phase151RecipientChannelAccess(grant bool) messenger.ChannelAccess {
	return messenger.ChannelAccess{
		Channel:                    &ent.Channel{ID: uuid.New()},
		Role:                       "member",
		IsMember:                   grant,
		HasActiveProjectMembership: grant,
		CanRead:                    grant,
		CanPost:                    grant,
		CanReply:                   grant,
		CanJoin:                    grant,
		CanFollow:                  grant,
		CanUnfollow:                grant,
		CanViewUnread:              grant,
		CanType:                    grant,
		CanEdit:                    grant,
		CanForward:                 grant,
		CanReact:                   grant,
		CanDeleteMessage:           grant,
		CanPin:                     grant,
		CanListPins:                grant,
		CanBookmark:                grant,
		CanPromote:                 grant,
	}
}

func phase151CapabilityValues(access messenger.ChannelAccess) map[string]bool {
	return map[string]bool{
		"is_member":                     access.IsMember,
		"has_active_project_membership": access.HasActiveProjectMembership,
		"can_read":                      access.CanRead,
		"can_post":                      access.CanPost,
		"can_reply":                     access.CanReply,
		"can_join":                      access.CanJoin,
		"can_follow":                    access.CanFollow,
		"can_unfollow":                  access.CanUnfollow,
		"can_view_unread":               access.CanViewUnread,
		"can_type":                      access.CanType,
		"can_edit":                      access.CanEdit,
		"can_forward":                   access.CanForward,
		"can_react":                     access.CanReact,
		"can_delete_message":            access.CanDeleteMessage,
		"can_pin":                       access.CanPin,
		"can_list_pins":                 access.CanListPins,
		"can_bookmark":                  access.CanBookmark,
		"can_promote":                   access.CanPromote,
	}
}

func phase151CapabilitiesJSON(t *testing.T, response any) map[string]bool {
	t.Helper()

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal recipient response: %v", err)
	}
	var payload struct {
		Capabilities map[string]bool `json:"capabilities"`
	}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("unmarshal recipient response: %v", err)
	}
	if payload.Capabilities == nil {
		t.Fatal("recipient response has no capabilities object")
	}
	return payload.Capabilities
}
