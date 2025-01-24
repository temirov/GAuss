package session

import (
	gsessions "github.com/gorilla/sessions"
)

var store *gsessions.CookieStore

// NewSession creates a new cookie store
func NewSession(secret []byte) {
	store = gsessions.NewCookieStore(secret)
	store.Options = &gsessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false, // Set to true in production
	}
}

// Store returns the global session store
func Store() *gsessions.CookieStore {
	if store == nil {
		panic("session store is nil")
	}
	return store
}
