package session

import (
	gsessions "github.com/gorilla/sessions"
)

var store *gsessions.CookieStore

// NewSession initializes the package-level cookie store with the given secret.
// It should be called once at application startup.
func NewSession(secret []byte) {
	store = gsessions.NewCookieStore(secret)
	store.Options = &gsessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false, // Set to true in production
	}
}

// Store returns the global session store previously created with NewSession.
// It panics if NewSession has not been called.
func Store() *gsessions.CookieStore {
	if store == nil {
		panic("session store is nil")
	}
	return store
}
