package dash

import (
	"github.com/gorilla/sessions"
	"github.com/temirov/GAuss/pkg/constants"
)

type Service struct {
	// Add any dashboard-specific dependencies
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetUserData(session *sessions.Session) map[string]interface{} {
	return map[string]interface{}{
		"Name":    session.Values[constants.SessionKeyUserName],
		"Email":   session.Values[constants.SessionKeyUserEmail],
		"Picture": session.Values[constants.SessionKeyUserPicture],
	}
}
