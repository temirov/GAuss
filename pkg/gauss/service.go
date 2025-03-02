package gauss

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type Service struct {
	config           *oauth2.Config
	localRedirectURL string
}

func NewService(clientID, clientSecret, googleOAuthBase, localRedirectURL string) (*Service, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("missing Google OAuth credentials")
	}

	baseURL, googleOAuthBaseErr := url.Parse(googleOAuthBase)
	if googleOAuthBaseErr != nil {
		return nil, errors.New("invalid Google OAuth base URL")
	}
	relativePath, _ := url.Parse(CallbackPath)
	redirectURL := baseURL.ResolveReference(relativePath)

	return &Service{
		config: &oauth2.Config{
			RedirectURL:  redirectURL.String(),
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"profile", "email"},
			Endpoint:     google.Endpoint,
		},
		localRedirectURL: localRedirectURL,
	}, nil
}

func (s *Service) GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *Service) GetUser(token *oauth2.Token) (*GoogleUser, error) {
	client := s.config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d", resp.StatusCode)
	}

	var user GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &user, nil
}
