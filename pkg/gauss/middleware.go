package gauss

import (
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"net/http"
)

func AuthMiddleware(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		webSession, _ := session.Store().Get(request, constants.SessionName)
		if webSession.Values[constants.SessionKeyUserEmail] == nil {
			http.Redirect(responseWriter, request, constants.LoginPath, http.StatusFound)
			return
		}
		nextHandler.ServeHTTP(responseWriter, request)
	})
}
