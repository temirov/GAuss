package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/gauss"
	"github.com/temirov/GAuss/pkg/session"
	"github.com/temirov/utils/system"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	DashboardPath = "/youtube"
	Root          = "/"
	appBase       = "http://localhost:8080/"
)

// Logger wraps standard logger with structured logging
type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

func (l *Logger) Error(msg string, err error, fields ...interface{}) {
	args := append([]interface{}{"ERROR", msg, "error", err}, fields...)
	l.Println(args...)
}

func (l *Logger) Info(msg string, fields ...interface{}) {
	args := append([]interface{}{"INFO", msg}, fields...)
	l.Println(args...)
}

func (l *Logger) Debug(msg string, fields ...interface{}) {
	args := append([]interface{}{"DEBUG", msg}, fields...)
	l.Println(args...)
}

var logger = NewLogger()

func main() {
	loginTemplateFlag := flag.String("template", "", "Path to custom login template (empty for default)")
	flag.Parse()

	clientSecret := system.GetEnvOrFail("SESSION_SECRET")
	googleClientID := system.GetEnvOrFail("GOOGLE_CLIENT_ID")
	googleClientSecret := system.GetEnvOrFail("GOOGLE_CLIENT_SECRET")

	session.NewSession([]byte(clientSecret))

	scopes := gauss.ScopeStrings([]gauss.Scope{gauss.ScopeProfile, gauss.ScopeEmail, gauss.ScopeYouTubeReadonly})
	authService, err := gauss.NewService(googleClientID, googleClientSecret, appBase, DashboardPath, scopes, *loginTemplateFlag)
	if err != nil {
		logger.Error("Failed to initialize auth service", err)
		os.Exit(1)
	}

	authHandlers, err := gauss.NewHandlers(authService)
	if err != nil {
		logger.Error("Failed to initialize handlers", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	authHandlers.RegisterRoutes(mux)

	templates, err := template.ParseGlob("examples/youtube_listing/templates/*.html")
	if err != nil {
		logger.Error("Failed to parse templates", err)
		os.Exit(1)
	}

	mux.Handle(DashboardPath, gauss.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderYouTube(w, r, authService, templates)
	})))

	mux.Handle(Root, gauss.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, DashboardPath, http.StatusFound)
	})))

	logger.Info("Server starting", "port", "8080")
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		logger.Error("Server failed to start", err)
		os.Exit(1)
	}
}

func renderYouTube(w http.ResponseWriter, r *http.Request, svc *gauss.Service, t *template.Template) {
	sess, err := session.Store().Get(r, constants.SessionName)
	if err != nil {
		logger.Error("Failed to get session", err, "remote_addr", r.RemoteAddr)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	tokJSON, ok := sess.Values[constants.SessionKeyOAuthToken].(string)
	if !ok {
		logger.Error("Missing OAuth token in session", nil, "remote_addr", r.RemoteAddr, "session_id", sess.ID)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokJSON), &token); err != nil {
		logger.Error("Failed to unmarshal OAuth token", err, "remote_addr", r.RemoteAddr, "token_length", len(tokJSON))
		http.Error(w, "Invalid authentication token", http.StatusInternalServerError)
		return
	}

	// Validate token before using it
	if token.AccessToken == "" {
		logger.Error("Empty access token", nil, "remote_addr", r.RemoteAddr, "token_type", token.TokenType)
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	logger.Debug("Creating YouTube service", "remote_addr", r.RemoteAddr, "token_type", token.TokenType, "has_refresh_token", token.RefreshToken != "")

	httpClient := svc.GetClient(r.Context(), &token)
	ytService, err := youtube.NewService(r.Context(), option.WithHTTPClient(httpClient))
	if err != nil {
		logger.Error("Failed to create YouTube service", err, "remote_addr", r.RemoteAddr)
		http.Error(w, "YouTube service unavailable", http.StatusInternalServerError)
		return
	}

	logger.Debug("Fetching YouTube channels", "remote_addr", r.RemoteAddr)
	channels, err := ytService.Channels.List([]string{"contentDetails"}).Mine(true).Do()
	if err != nil {
		logger.Error("Failed to fetch YouTube channels", err, "remote_addr", r.RemoteAddr, "token_expired", token.Expiry.Before(time.Now()))

		// Check if it's a token expiry issue
		if token.RefreshToken == "" {
			logger.Error("No refresh token available for expired token", nil, "remote_addr", r.RemoteAddr)
			http.Redirect(w, r, "/logout", http.StatusFound)
			return
		}

		http.Error(w, "Failed to access YouTube data", http.StatusInternalServerError)
		return
	}

	if len(channels.Items) == 0 {
		logger.Info("No YouTube channels found for user", "remote_addr", r.RemoteAddr)
		http.Error(w, "No YouTube channel found", http.StatusNotFound)
		return
	}

	uploads := channels.Items[0].ContentDetails.RelatedPlaylists.Uploads
	logger.Debug("Fetching playlist items", "remote_addr", r.RemoteAddr, "uploads_playlist", uploads)

	vids, err := ytService.PlaylistItems.List([]string{"snippet"}).PlaylistId(uploads).MaxResults(10).Do()
	if err != nil {
		logger.Error("Failed to fetch playlist items", err, "remote_addr", r.RemoteAddr, "playlist_id", uploads)
		http.Error(w, "Failed to load videos", http.StatusInternalServerError)
		return
	}

	logger.Info("Successfully loaded YouTube videos", "remote_addr", r.RemoteAddr, "video_count", len(vids.Items))

	if err := t.ExecuteTemplate(w, "youtube_videos.html", vids.Items); err != nil {
		logger.Error("Failed to execute template", err, "remote_addr", r.RemoteAddr)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
}
