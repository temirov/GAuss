package constants

// Paths and keys used throughout GAuss.

const (
	// LoginPath is the route that serves the login page.
	LoginPath = "/login"
	// GoogleAuthPath starts the OAuth2 flow with Google.
	GoogleAuthPath = "/auth/google"
	// CallbackPath receives the OAuth2 redirect from Google.
	CallbackPath = "/auth/google/callback"
	// LogoutPath clears the user session.
	LogoutPath = "/logout"
	// TemplatesPath points to embedded login templates.
	TemplatesPath = "templates/*.html"
	// DefaultTemplateName is the embedded login template name.
	DefaultTemplateName = "login.html"

	// SessionKeyUserEmail stores the logged-in user's email in the session.
	SessionKeyUserEmail = "user_email"
	// SessionKeyUserName stores the logged-in user's display name.
	SessionKeyUserName = "user_name"
	// SessionKeyUserPicture stores the profile image URL.
	SessionKeyUserPicture = "user_picture"

	// SessionName is the cookie name used for sessions.
	SessionName = "gauss_session"
)
