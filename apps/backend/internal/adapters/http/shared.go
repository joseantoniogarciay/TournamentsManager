// Package http exposes the API HTTP adapters.
package http

import (
	"regexp"
	"time"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const (
	usernameAvailabilityLimit  = 30
	usernameAvailabilityWindow = time.Minute
	userSearchLimit            = 60
	registrationLimit          = 5
	suggestionLimit            = 3
	suggestionWindow           = time.Hour
)
