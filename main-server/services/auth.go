package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"main-server/models"

	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthService struct {
	clientId     string
	domain       string
	redirectURL  string
	provider     *oidc.Provider
	oauth2Config oauth2.Config
}

type ClaimsData struct {
	AccessToken string
	IDToken     string
	Claims      jwt.MapClaims
	User        *models.User
}

func NewAuthService(clientId, clientSecret, redirectURL, domain, region, userPoolId string) (*AuthService, error) {
	var err error

	issuerURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", region, userPoolId)

	// Initialize OIDC provider
	provider, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	// Set up OAuth2 config
	oauth2Config := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "phone", "email"},
	}

	return &AuthService{
		provider:     provider,
		oauth2Config: oauth2Config,
		clientId:     clientId,
		domain:       domain,
		redirectURL:  redirectURL,
	}, nil
}

func (a *AuthService) GetLoginURL(customState ...string) (string, string, error) {
	var state string
	var err error

	if len(customState) > 0 && customState[0] != "" {
		// Use provided state
		state = customState[0]
	} else {
		// Generate secure state
		state, err = a.generateSecureState()
		if err != nil {
			return "", "", err
		}
	}

	url := a.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	return url, state, nil
}

func (a *AuthService) HandleCallback(code, state string, storedState string) (*ClaimsData, error) {
	// Verify state parameter
	if state != storedState {
		return nil, fmt.Errorf("invalid state parameter")
	}

	ctx := context.Background()

	// Exchange the authorization code for tokens
	rawToken, err := a.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := rawToken.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in response")
	}

	// Verify the ID token
	idToken, err := a.provider.Verifier(&oidc.Config{
		ClientID: a.oauth2Config.ClientID,
	}).Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}

	// Parse claims from ID token
	var claims jwt.MapClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	// Extract user info
	user := &models.User{}
	if err := idToken.Claims(user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &ClaimsData{
		AccessToken: rawToken.AccessToken,
		IDToken:     rawIDToken,
		Claims:      claims,
		User:        user,
	}, nil
}

func (a *AuthService) GetLogoutURL(returnToURL string) string {
	logoutURL := fmt.Sprintf("https://%s/logout?client_id=%s&logout_uri=%s",
		a.domain,
		a.clientId,
		returnToURL)
	return logoutURL
}

func (a *AuthService) generateSecureState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Session helpers
func (a *AuthService) StoreUserSession(c echo.Context, claimsData *ClaimsData) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Values["user_id"] = claimsData.User
	sess.Values["email"] = claimsData.User.Email
	sess.Values["access_token"] = claimsData.AccessToken
	sess.Values["id_token"] = claimsData.IDToken
	sess.Values["user_tier"] = claimsData.User.Role
	sess.Values["company_id"] = claimsData.User.CompanyID
	sess.Values["authenticated"] = true

	return sess.Save(c.Request(), c.Response())
}

func (a *AuthService) ClearUserSession(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1

	return sess.Save(c.Request(), c.Response())
}

func (a *AuthService) GetCurrentUser(c echo.Context) (*models.User, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, err
	}

	authenticated, ok := sess.Values["authenticated"].(bool)
	if !ok || !authenticated {
		return nil, fmt.Errorf("user not authenticated")
	}

	return &models.User{
		ID:        sess.Values["id"].(string),
		Email:     sess.Values["email"].(string),
		Role:      sess.Values["role"].(string),
		CompanyID: sess.Values["company_id"].(string),
	}, nil
}
