package dash

import (
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"html/template"
	"net/http"
)

type Handlers struct {
	service   *Service
	templates *template.Template
}

func NewHandlers(service *Service, templates *template.Template) *Handlers {
	return &Handlers{
		service:   service,
		templates: templates,
	}
}

func (handlers *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	webSession, _ := session.Store().Get(r, constants.SessionName)
	data := handlers.service.GetUserData(webSession)
	handlers.templates.ExecuteTemplate(w, "dashboard.html", data)
}

func (handlers *Handlers) RegisterRoutes(mux *http.ServeMux, path string) {
	mux.HandleFunc(path, handlers.Dashboard)
	// Add other dashboard routes
}
