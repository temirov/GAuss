// Package dash contains handlers and services for rendering an authenticated
// dashboard. It relies on a GAuss session to retrieve user information and
// demonstrates how a calling application can serve custom pages once the user
// has authenticated through GAuss.
//
// The package exposes Handlers for wiring into an http.ServeMux and a Service
// type that extracts profile data from the current session.
package dash
