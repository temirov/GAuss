package gauss

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gorilla/sessions"
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"golang.org/x/oauth2"
)

//go:embed templates/*.html
var templatesFileSystem embed.FS

// Handlers bundles the GAuss service, session store, and HTML templates used for authentication.
type Handlers struct {
	service   *Service
	store     *sessions.CookieStore
	templates *template.Template
}

// NewHandlers creates a new Handlers instance using the GAuss service.
// It uses the service's LoginTemplate field: if it is non-empty, the template
// is loaded from the provided file path; otherwise, the default embedded templates
// (as specified by constants.TemplatesPath) are parsed.
func NewHandlers(serviceInstance *Service) (*Handlers, error) {
	var (
		parsedTemplates *template.Template
		err             error
	)
	if serviceInstance.LoginTemplate != "" {
		parsedTemplates, err = template.ParseFiles(serviceInstance.LoginTemplate)
	} else {
		parsedTemplates, err = template.ParseFS(templatesFileSystem, constants.TemplatesPath)
	}
	if err != nil {
		return nil, err
	}

	cookieStore := session.Store()

	return &Handlers{
		service:   serviceInstance,
		store:     cookieStore,
		templates: parsedTemplates,
	}, nil
}

// RegisterRoutes registers the authentication routes.
func (handlersInstance *Handlers) RegisterRoutes(httpMux *http.ServeMux) *http.ServeMux {
	httpMux.HandleFunc(constants.LoginPath, handlersInstance.loginHandler)
	httpMux.HandleFunc(constants.GoogleAuthPath, handlersInstance.Login)
	httpMux.HandleFunc(constants.CallbackPath, handlersInstance.Callback)
	httpMux.HandleFunc(constants.LogoutPath, handlersInstance.Logout)

	return httpMux
}

// loginHandler renders the login page.
// It looks up the template by the base name of the custom template file if provided,
// or defaults to constants.DefaultTemplateName.
func (handlersInstance *Handlers) loginHandler(responseWriter http.ResponseWriter, request *http.Request) {
	dataMap := map[string]interface{}{
		"error": request.URL.Query().Get("error"),
	}

	var templateName string
	if handlersInstance.service.LoginTemplate != "" {
		templateName = filepath.Base(handlersInstance.service.LoginTemplate)
	} else {
		templateName = constants.DefaultTemplateName
	}

	tmpl := handlersInstance.templates.Lookup(templateName)
	if tmpl == nil {
		http.Error(responseWriter, "Login template not found", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(responseWriter, dataMap); err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Login handles the OAuth2 login process.
func (handlersInstance *Handlers) Login(responseWriter http.ResponseWriter, request *http.Request) {
	stateValue, stateError := handlersInstance.service.GenerateState()
	if stateError != nil {
		log.Printf("Failed to generate state: %v", stateError)
		http.Error(responseWriter, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	webSession, _ := handlersInstance.store.Get(request, constants.SessionName)
	webSession.Values["oauth_state"] = stateValue
	if sessionSaveError := webSession.Save(request, responseWriter); sessionSaveError != nil {
		log.Printf("Failed to save session: %v", sessionSaveError)
		http.Error(responseWriter, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	authorizationURL := handlersInstance.service.config.AuthCodeURL(stateValue, oauth2.SetAuthURLParam("prompt", "select_account"))
	http.Redirect(responseWriter, request, authorizationURL, http.StatusFound)
}

// Callback handles the OAuth2 callback, validates the state and code, and logs the user in.
func (handlersInstance *Handlers) Callback(responseWriter http.ResponseWriter, request *http.Request) {
	webSession, _ := handlersInstance.store.Get(request, constants.SessionName)
	storedStateValue, stateOk := webSession.Values["oauth_state"].(string)
	if !stateOk {
		log.Println("Missing state in session")
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=missing_state", http.StatusFound)
		return
	}

	receivedStateValue := request.URL.Query().Get("state")
	if storedStateValue != receivedStateValue {
		log.Printf("State mismatch: stored %s vs received %s", storedStateValue, receivedStateValue)
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=invalid_state", http.StatusFound)
		return
	}

	authorizationCode := request.URL.Query().Get("code")
	if authorizationCode == "" {
		log.Println("Missing authorization code")
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=missing_code", http.StatusFound)
		return
	}

	oauthToken, tokenExchangeError := handlersInstance.service.config.Exchange(request.Context(), authorizationCode)
	if tokenExchangeError != nil {
		log.Printf("Token exchange failed: %v", tokenExchangeError)
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=token_exchange_failed", http.StatusFound)
		return
	}

	googleUser, getUserError := handlersInstance.service.GetUser(oauthToken)
	if getUserError != nil {
		log.Printf("Failed to get user info: %v", getUserError)
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=user_info_failed", http.StatusFound)
		return
	}

	webSession.Values[constants.SessionKeyUserEmail] = googleUser.Email
	webSession.Values[constants.SessionKeyUserName] = googleUser.Name
	webSession.Values[constants.SessionKeyUserPicture] = googleUser.Picture
	if sessionSaveError := webSession.Save(request, responseWriter); sessionSaveError != nil {
		log.Printf("Failed to save user session: %v", sessionSaveError)
		http.Redirect(responseWriter, request, constants.LoginPath+"?error=session_save_failed", http.StatusFound)
		return
	}

	http.Redirect(responseWriter, request, handlersInstance.service.localRedirectURL, http.StatusFound)
}

// Logout clears the user session and redirects to the login page.
func (handlersInstance *Handlers) Logout(responseWriter http.ResponseWriter, request *http.Request) {
	webSession, _ := handlersInstance.store.Get(request, constants.SessionName)
	webSession.Options.MaxAge = -1
	if webSessionSaveError := webSession.Save(request, responseWriter); webSessionSaveError != nil {
		http.Error(responseWriter, webSessionSaveError.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(responseWriter, request, constants.LoginPath, http.StatusFound)
}
