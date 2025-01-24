package gauss

import (
	"github.com/temirov/GAuss/pkg/constants"
	"github.com/temirov/GAuss/pkg/session"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webSession, _ := session.Store().Get(r, constants.SessionName)
		if webSession.Values["user_email"] == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
