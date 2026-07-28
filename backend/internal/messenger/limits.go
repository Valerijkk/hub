package messenger

import "errors"

// ErrInviteUsesExhausted — атомарный IncrementInviteUses не затронул строку
// (use_count уже >= max_uses или invite исчез).
var ErrInviteUsesExhausted = errors.New("invite use limit exhausted")

// Hard caps для Ent .All() без курсора (A2W3-3-002 / W3-A3-006).
const (
	listMembersCap          = 500
	listPinsCap             = 200
	listInvitesCap          = 100
	listBookmarksCap        = 500
	listVotesCap            = 2000
	listProjectMembersCap   = 500
	channelMemberIDsCap     = 500
	userChannelMembershipCap = 500
)

// TTL ключей unread:{userID} в Redis (A2W3-3-003).
const unreadKeyTTLSeconds = 30 * 24 * 60 * 60 // 30 дней
