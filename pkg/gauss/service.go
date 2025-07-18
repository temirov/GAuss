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

// userInfoEndpoint specifies the URL used to retrieve profile information from
// Google. It is a variable rather than a constant so tests can replace it with
// a mock server endpoint.
var userInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"

// GoogleUser represents a user profile retrieved from Google.
type GoogleUser struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Service encapsulates OAuth2 configuration and redirection settings used by
// GAuss. It generates the authorization URL, validates callbacks and provides
// helper methods for retrieving the authenticated user's profile.
//
// The LoginTemplate field, if non-empty, specifies the HTML template filename
// to be used for the login page instead of the embedded "login.html".
type Service struct {
	config           *oauth2.Config
	localRedirectURL string
	LoginTemplate    string
}

// NewService initializes a Service with Google OAuth credentials and the local
// redirect URL where authenticated users will be sent after logging in.
// googleOAuthBase should point to the publicly reachable URL of your GAuss
// application (e.g. "http://localhost:8080"). customLoginTemplate may specify
// a login template file to override the default.
func NewService(clientID string, clientSecret string, googleOAuthBase string, localRedirectURL string, scopes []string, customLoginTemplate string) (*Service, error) {
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("missing Google OAuth credentials")
	}

	baseURL, googleOAuthBaseErr := url.Parse(googleOAuthBase)
	if googleOAuthBaseErr != nil {
		return nil, errors.New("invalid Google OAuth base URL")
	}
	relativePath, _ := url.Parse(constants.CallbackPath)
	redirectURL := baseURL.ResolveReference(relativePath)

	if len(scopes) == 0 {
		scopes = ScopeStrings(DefaultScopes)
	}

	return &Service{
		config: &oauth2.Config{
			RedirectURL:  redirectURL.String(),
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
		localRedirectURL: localRedirectURL,
		LoginTemplate:    customLoginTemplate,
	}, nil
}

// GenerateState returns a cryptographically secure random string that is used
// as the OAuth2 state parameter to protect against cross-site request forgery.
func (serviceInstance *Service) GenerateState() (string, error) {
	randomBytes := make([]byte, 32)
	_, readError := rand.Read(randomBytes)
	if readError != nil {
		return "", fmt.Errorf("failed to generate state: %w", readError)
	}
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

// GetUser contacts Google's userinfo endpoint to retrieve the profile
// associated with the provided OAuth2 token.
func (serviceInstance *Service) GetUser(oauthToken *oauth2.Token) (*GoogleUser, error) {
	httpClient := serviceInstance.config.Client(context.Background(), oauthToken)
	httpResponse, httpError := httpClient.Get(userInfoEndpoint)
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
