package gauss

import (
	"embed"
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"golang.org/x/oauth2"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

//go:embed templates/*.html
var templatesFileSystem embed.FS

type Handlers struct {
	service   *Service
	store     *sessions.CookieStore
	templates *template.Template
}

func NewHandlers(serviceInstance *Service) (*Handlers, error) {
	parsedTemplates, parseError := template.ParseFS(templatesFileSystem, constants.TemplatesPath)
	if parseError != nil {
		return nil, parseError
	}

	cookieStore := session.Store()

	return &Handlers{
		service:   serviceInstance,
		store:     cookieStore,
		templates: parsedTemplates,
	}, nil
}

func (handlersInstance *Handlers) RegisterRoutes(httpMux *http.ServeMux) *http.ServeMux {
	httpMux.HandleFunc(constants.LoginPath, handlersInstance.loginHandler)
	httpMux.HandleFunc(constants.GoogleAuthPath, handlersInstance.Login)
	httpMux.HandleFunc(constants.CallbackPath, handlersInstance.Callback)
	httpMux.HandleFunc(constants.LogoutPath, handlersInstance.Logout)

	return httpMux
}

func (handlersInstance *Handlers) loginHandler(responseWriter http.ResponseWriter, request *http.Request) {
	dataMap := map[string]interface{}{
		"error": request.URL.Query().Get("error"),
	}
	templateParsingError := handlersInstance.templates.ExecuteTemplate(responseWriter, "login.html", dataMap)
	if templateParsingError != nil {
		http.Error(responseWriter, templateParsingError.Error(), http.StatusInternalServerError)
		return
	}
}

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

func (handlersInstance *Handlers) Logout(responseWriter http.ResponseWriter, request *http.Request) {
	webSession, _ := handlersInstance.store.Get(request, constants.SessionName)
	webSession.Options.MaxAge = -1
	webSessionSaveError := webSession.Save(request, responseWriter)
	if webSessionSaveError != nil {
		http.Error(responseWriter, webSessionSaveError.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(responseWriter, request, constants.LoginPath, http.StatusFound)
}
