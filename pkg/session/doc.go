// Package session wraps gorilla/sessions to provide a global cookie store used
// by GAuss. Call NewSession with your secret key at startup to initialize the
// store and then use Store to retrieve it whenever a handler needs access to the
// session. The package is intentionally small so that other packages can share
// session management without having to configure gorilla/sessions directly.
package session
