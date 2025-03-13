package gauss

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/temirov/GAuss/pkg/constants"
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

func NewService(clientID string, clientSecret string, googleOAuthBase string, localRedirectURL string) (*Service, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("missing Google OAuth credentials")
	}

	baseURL, googleOAuthBaseErr := url.Parse(googleOAuthBase)
	if googleOAuthBaseErr != nil {
		return nil, errors.New("invalid Google OAuth base URL")
	}
	relativePath, _ := url.Parse(constants.CallbackPath)
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

func (serviceInstance *Service) GenerateState() (string, error) {
	randomBytes := make([]byte, 32)
	_, readError := rand.Read(randomBytes)
	if readError != nil {
		return "", fmt.Errorf("failed to generate state: %w", readError)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

func (serviceInstance *Service) GetUser(oauthToken *oauth2.Token) (*GoogleUser, error) {
	httpClient := serviceInstance.config.Client(context.Background(), oauthToken)
	httpResponse, httpError := httpClient.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if httpError != nil {
		return nil, fmt.Errorf("failed to get user info: %w", httpError)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google API returned status %d", httpResponse.StatusCode)
	}

	var user GoogleUser
	if decodeError := json.NewDecoder(httpResponse.Body).Decode(&user); decodeError != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", decodeError)
	}

	return &user, nil
}
