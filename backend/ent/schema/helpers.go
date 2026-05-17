package schema

import (
	"regexp"
	"time"
)

// keyRegexp validates project keys: starts with uppercase letter, followed by 1-9 uppercase letters or digits.
var keyRegexp = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)

// nowUTC returns the current time in UTC. It is the canonical default/update
// hook for timestamp fields in Ent schemas, ensuring Postgres receives
// UTC-normalised values regardless of the server's local timezone.
func nowUTC() time.Time { return time.Now().UTC() }
