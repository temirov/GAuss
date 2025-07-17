// Package gauss implements Google OAuth2 authentication and session management.
//
// It exposes a Service type that configures the OAuth2 client, generates state
// parameters, and retrieves user information from Google. Handlers provides a
// ready-to-use set of HTTP handlers that mount the login, callback, and logout
// routes on a ServeMux. The package also offers AuthMiddleware to protect
// application endpoints by ensuring a valid GAuss session is present.
//
// Applications can embed this package to replace custom Google authentication
// flows. Initialize a Service with your OAuth credentials, create Handlers, and
// register the routes with your own mux. Protected routes should be wrapped with
// AuthMiddleware to require a logged in user.
package gauss
