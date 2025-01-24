package main

import (
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/dash"
	"github.com/temirov/GAuss/pkg/session"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/temirov/GAuss/pkg/gauss"
)

const (
	DashboardPath = "/dashboard"
	Root          = "/"
)

func GetEnvOrFail(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s environment variable not set", name)
	}
	return value
}

func main() {
	clientSecret := GetEnvOrFail("SESSION_SECRET")
	googleClientID := GetEnvOrFail("GOOGLE_CLIENT_ID")
	googleClientSecret := GetEnvOrFail("GOOGLE_CLIENT_SECRET")

	session.NewSession([]byte(clientSecret))

	authService, err := gauss.NewService(googleClientID, googleClientSecret, DashboardPath)
	if err != nil {
		log.Fatalf("Failed to initialize auth service: %v", err)
	}

	authHandlers, err := gauss.NewHandlers(authService)
	if err != nil {
		log.Fatalf("Failed to initialize handlers: %v", err)
	}

	// Set up routing
	mux := http.NewServeMux()

	// Auth routes (unprotected)
	authHandlers.RegisterRoutes(mux)

	// Initialize dashboard service and handlers
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal(err)
	}
	dashService := dash.NewService()
	dashHandlers := dash.NewHandlers(dashService, templates)

	mux.Handle(DashboardPath, gauss.AuthMiddleware(http.HandlerFunc(dashHandlers.Dashboard)))

	// Register root handler with middleware
	mux.Handle(Root, gauss.AuthMiddleware(http.HandlerFunc(rootHandler)))

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	webSession, _ := session.Store().Get(r, constants.SessionName)
	if webSession.Values["user_email"] != nil {
		// User is logged in, redirect to dashboard
		http.Redirect(w, r, DashboardPath, http.StatusFound)
		return
	}
	// If not logged in, the middleware will handle the redirect to login
	http.NotFound(w, r)
}
