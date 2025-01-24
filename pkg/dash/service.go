package dash

import (
	"github.com/gorilla/sessions"
)

type Service struct {
	// Add any dashboard-specific dependencies
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetUserData(session *sessions.Session) map[string]interface{} {
	return map[string]interface{}{
		"Name":    session.Values["user_name"],
		"Email":   session.Values["user_email"],
		"Picture": session.Values["user_picture"],
	}
}
