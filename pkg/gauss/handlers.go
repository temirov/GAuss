package gauss

import (
	"embed"
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	LoginPath      = "/login"
	GoogleAuthPath = "/auth/google"
	CallbackPath   = "/auth/google/callback"
	LogoutPath     = "/logout"
	TemplatesPath  = "templates/*.html"
)

//go:embed templates/*.html
var templatesFS embed.FS

type Handlers struct {
	service   *Service
	store     *sessions.CookieStore
	templates *template.Template
}

func NewHandlers(service *Service) (*Handlers, error) {
	templates, err := template.ParseFS(templatesFS, TemplatesPath)
	if err != nil {
		return nil, err
	}

	store := session.Store()

	return &Handlers{
		service:   service,
		store:     store,
		templates: templates,
	}, nil
}

func (handlers *Handlers) RegisterRoutes(mux *http.ServeMux) *http.ServeMux {
	mux.HandleFunc(LoginPath, handlers.loginHandler)
	mux.HandleFunc(GoogleAuthPath, handlers.Login)
	mux.HandleFunc(CallbackPath, handlers.Callback)
	mux.HandleFunc(LogoutPath, handlers.Logout)

	return mux
}

func (handlers *Handlers) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, LoginPath, http.StatusFound)
}

func (handlers *Handlers) loginHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"error": r.URL.Query().Get("error"),
	}
	handlers.templates.ExecuteTemplate(w, "login.html", data)
}

func (handlers *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	state, err := handlers.service.GenerateState()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	webSession, _ := handlers.store.Get(r, constants.SessionName)
	webSession.Values["oauth_state"] = state
	if sessionError := webSession.Save(r, w); sessionError != nil {
		log.Printf("Failed to save session: %v", sessionError)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	authURL := handlers.service.config.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (handlers *Handlers) Callback(w http.ResponseWriter, request *http.Request) {
	webSession, _ := handlers.store.Get(request, constants.SessionName)
	storedState, ok := webSession.Values["oauth_state"].(string)
	if !ok {
		log.Println("Missing state in session")
		http.Redirect(w, request, LoginPath+"?error=missing_state", http.StatusFound)
		return
	}

	queryState := request.URL.Query().Get("state")
	if storedState != queryState {
		log.Printf("State mismatch: stored %s vs received %s", storedState, queryState)
		http.Redirect(w, request, LoginPath+"?error=invalid_state", http.StatusFound)
		return
	}

	code := request.URL.Query().Get("code")
	if code == "" {
		log.Println("Missing authorization code")
		http.Redirect(w, request, LoginPath+"?error=missing_code", http.StatusFound)
		return
	}

	token, err := handlers.service.config.Exchange(request.Context(), code)
	if err != nil {
		log.Printf("Token exchange failed: %v", err)
		http.Redirect(w, request, LoginPath+"?error=token_exchange_failed", http.StatusFound)
		return
	}

	user, err := handlers.service.GetUser(token)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		http.Redirect(w, request, LoginPath+"?error=user_info_failed", http.StatusFound)
		return
	}

	webSession.Values["user_email"] = user.Email
	webSession.Values["user_name"] = user.Name
	webSession.Values["user_picture"] = user.Picture
	if sessionSaveError := webSession.Save(request, w); sessionSaveError != nil {
		log.Printf("Failed to save user session: %v", sessionSaveError)
		http.Redirect(w, request, LoginPath+"?error=session_save_failed", http.StatusFound)
		return
	}

	http.Redirect(w, request, handlers.service.localRedirectURL, http.StatusFound)
}

func (handlers *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	webSession, _ := handlers.store.Get(r, constants.SessionName)
	webSession.Options.MaxAge = -1
	webSession.Save(r, w)
	http.Redirect(w, r, LoginPath, http.StatusFound)
}
