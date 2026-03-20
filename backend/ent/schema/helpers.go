package schema

import "regexp"

// keyRegexp validates project keys: starts with uppercase letter, followed by 1-9 uppercase letters or digits.
var keyRegexp = regexp.MustCompile(`^[A-Z][A-Z0-9]{1,9}$`)
