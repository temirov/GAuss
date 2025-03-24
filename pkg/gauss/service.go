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

	"github.com/temirov/GAuss/pkg/constants"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUser represents a user profile retrieved from Google.
type GoogleUser struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Service encapsulates OAuth2 configuration and redirection settings.
// The LoginTemplate field, if non-empty, specifies the HTML template filename
// to be used for login instead of the default "login.html".
type Service struct {
	config           *oauth2.Config
	localRedirectURL string
	LoginTemplate    string
}

// NewService initializes a new Service instance.
// The customLoginTemplate parameter is the filename (e.g. "custom_login.html") to be used for the login page.
// Pass an empty string to use the default template.
func NewService(clientID string, clientSecret string, googleOAuthBase string, localRedirectURL string, customLoginTemplate string) (*Service, error) {
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
		LoginTemplate:    customLoginTemplate,
	}, nil
}

// GenerateState creates a new random state string.
func (serviceInstance *Service) GenerateState() (string, error) {
	randomBytes := make([]byte, 32)
	_, readError := rand.Read(randomBytes)
	if readError != nil {
		return "", fmt.Errorf("failed to generate state: %w", readError)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

// GetUser retrieves the Google user profile for the given OAuth2 token.
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
