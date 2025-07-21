package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

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
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	authHandlers, err := gauss.NewHandlers(authService)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	mux := http.NewServeMux()
	authHandlers.RegisterRoutes(mux)

	templates, err := template.ParseGlob("examples/youtube_listing/templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle(DashboardPath, gauss.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderYouTube(w, r, authService, templates)
	})))

	mux.Handle(Root, gauss.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, DashboardPath, http.StatusFound)
	})))

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", mux))
}

func renderYouTube(w http.ResponseWriter, r *http.Request, svc *gauss.Service, t *template.Template) {
	sess, _ := session.Store().Get(r, constants.SessionName)
	tokJSON, ok := sess.Values[constants.SessionKeyOAuthToken].(string)
	if !ok {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	var token oauth2.Token
	if err := json.Unmarshal([]byte(tokJSON), &token); err != nil {
		http.Error(w, "invalid token", http.StatusInternalServerError)
		return
	}

	httpClient := svc.GetClient(r.Context(), &token)
	ytService, err := youtube.NewService(r.Context(), option.WithHTTPClient(httpClient))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	channels, err := ytService.Channels.List([]string{"contentDetails"}).Mine(true).Do()
	if err != nil || len(channels.Items) == 0 {
		http.Error(w, "failed to get channel", http.StatusInternalServerError)
		return
	}

	uploads := channels.Items[0].ContentDetails.RelatedPlaylists.Uploads
	vids, err := ytService.PlaylistItems.List([]string{"snippet"}).PlaylistId(uploads).MaxResults(10).Do()
	if err != nil {
		http.Error(w, "failed to list videos", http.StatusInternalServerError)
		return
	}

	t.ExecuteTemplate(w, "youtube_videos.html", vids.Items)
}
