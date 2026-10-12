// Package sessions defines the projection and operations for account sessions.
package sessions

import "context"

// Client contains advisory categories, never evidence for authorization.
type Client struct {
	Platform string
	Kind     string
}

// Item is the public projection of a stable session family.
type Item struct {
	ID         string `json:"id"`
	CreatedAt  string `json:"createdAt"`
	LastSeenAt string `json:"lastSeenAt"`
	Platform   string `json:"platform"`
	Client     string `json:"client"`
	Current    bool   `json:"current"`
}

// Page bounds work and payload size for accounts with many sessions.
type Page struct {
	Items      []Item  `json:"items"`
	NextCursor *string `json:"nextCursor"`
}

// Repository implements ownership and atomic revocation at the persistence boundary.
type Repository interface {
	ListSessions(context.Context, string, string) (Page, error)
	RevokeOtherSessions(context.Context, string, string) error
}
