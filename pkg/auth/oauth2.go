package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OAuth2Auth implements OAuth2 authentication with support for multiple flows.
type OAuth2Auth struct {
	config       *OAuth2Config
	oauth2Config *oauth2.Config
	ccConfig     *clientcredentials.Config
	pkceVerifier string
}

// NewOAuth2Auth creates a new OAuth2 authenticator.
func NewOAuth2Auth(config *OAuth2Config) (*OAuth2Auth, error) {
	if config == nil {
		return nil, fmt.Errorf("OAuth2 config is required")
	}

	auth := &OAuth2Auth{
		config: config,
	}

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	// Initialize the appropriate OAuth2 configuration
	if err := auth.initConfig(); err != nil {
		return nil, err
	}

	return auth, nil
}

// Type returns the authentication type.
func (o *OAuth2Auth) Type() AuthType {
	return AuthTypeOAuth2
}

// Authenticate performs OAuth2 authentication based on the configured flow.
func (o *OAuth2Auth) Authenticate(ctx context.Context) (*Token, error) {
	switch o.config.Flow {
	case OAuth2FlowAuthorizationCode:
		return o.authenticateAuthorizationCode(ctx)
	case OAuth2FlowClientCredentials:
		return o.authenticateClientCredentials(ctx)
	case OAuth2FlowPassword:
		return o.authenticatePassword(ctx)
	case OAuth2FlowDeviceCode:
		return o.authenticateDeviceCode(ctx)
	default:
		return nil, fmt.Errorf("unsupported OAuth2 flow: %s", o.config.Flow)
	}
}

// RefreshToken refreshes an OAuth2 token.
func (o *OAuth2Auth) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token == nil || token.RefreshToken == "" {
		return nil, fmt.Errorf("refresh token not available")
	}

	// Create token source from refresh token
	tok := &oauth2.Token{
		RefreshToken: token.RefreshToken,
		Expiry:       token.ExpiresAt,
	}

	tokenSource := o.oauth2Config.TokenSource(ctx, tok)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return convertOAuth2Token(newToken), nil
}

// GetHeaders returns HTTP headers for OAuth2 authentication.
func (o *OAuth2Auth) GetHeaders(token *Token) map[string]string {
	if token == nil || token.AccessToken == "" {
		return nil
	}

	tokenType := token.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return map[string]string{
		"Authorization": tokenType + " " + token.AccessToken,
	}
}

// Validate checks if the OAuth2 configuration is valid.
func (o *OAuth2Auth) Validate() error {
	if o.config == nil {
		return fmt.Errorf("OAuth2 config is required")
	}

	if o.config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if o.config.Flow == "" {
		return fmt.Errorf("flow is required")
	}

	// Validate flow-specific requirements
	switch o.config.Flow {
	case OAuth2FlowAuthorizationCode:
		if o.config.AuthURL == "" {
			return fmt.Errorf("auth_url is required for authorization_code flow")
		}
		if o.config.TokenURL == "" {
			return fmt.Errorf("token_url is required for authorization_code flow")
		}
	case OAuth2FlowClientCredentials:
		if o.config.TokenURL == "" {
			return fmt.Errorf("token_url is required for client_credentials flow")
		}
		if o.config.ClientSecret == "" {
			return fmt.Errorf("client_secret is required for client_credentials flow")
		}
	case OAuth2FlowPassword:
		if o.config.TokenURL == "" {
			return fmt.Errorf("token_url is required for password flow")
		}
		if o.config.Username == "" || o.config.Password == "" {
			return fmt.Errorf("username and password are required for password flow")
		}
	case OAuth2FlowDeviceCode:
		if o.config.DeviceCodeURL == "" {
			return fmt.Errorf("device_code_url is required for device_code flow")
		}
		if o.config.TokenURL == "" {
			return fmt.Errorf("token_url is required for device_code flow")
		}
	default:
		return fmt.Errorf("unsupported OAuth2 flow: %s", o.config.Flow)
	}

	return nil
}

// initConfig initializes the OAuth2 configuration based on the flow.
func (o *OAuth2Auth) initConfig() error {
	switch o.config.Flow {
	case OAuth2FlowClientCredentials:
		o.ccConfig = &clientcredentials.Config{
			ClientID:       o.config.ClientID,
			ClientSecret:   o.config.ClientSecret,
			TokenURL:       o.config.TokenURL,
			Scopes:         o.config.Scopes,
			EndpointParams: convertEndpointParams(o.config.EndpointParams),
		}
	default:
		endpoint := oauth2.Endpoint{
			AuthURL:  o.config.AuthURL,
			TokenURL: o.config.TokenURL,
		}
		if o.config.DeviceCodeURL != "" {
			endpoint.DeviceAuthURL = o.config.DeviceCodeURL
		}

		o.oauth2Config = &oauth2.Config{
			ClientID:     o.config.ClientID,
			ClientSecret: o.config.ClientSecret,
			RedirectURL:  o.config.RedirectURL,
			Scopes:       o.config.Scopes,
			Endpoint:     endpoint,
		}
	}

	return nil
}

// authenticateAuthorizationCode performs the authorization code flow.
func (o *OAuth2Auth) authenticateAuthorizationCode(ctx context.Context) (*Token, error) {
	// Start local server to receive callback
	server, callbackChan, err := o.startCallbackServer()
	if err != nil {
		return nil, err
	}
	defer func() { _ = server.Close() }()

	// Build authorization URL
	authURL := o.buildAuthURL()

	fmt.Printf("Please visit this URL to authorize:\n\n%s\n\n", authURL)
	fmt.Println("Waiting for authorization...")

	// Wait for callback or timeout
	select {
	case code := <-callbackChan:
		if code == "" {
			return nil, fmt.Errorf("authorization failed")
		}

		// Exchange code for token
		var opts []oauth2.AuthCodeOption
		if o.config.PKCE && o.pkceVerifier != "" {
			opts = append(opts, oauth2.VerifierOption(o.pkceVerifier))
		}

		token, err := o.oauth2Config.Exchange(ctx, code, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to exchange code for token: %w", err)
		}

		return convertOAuth2Token(token), nil

	case <-ctx.Done():
		return nil, fmt.Errorf("authorization cancelled")
	case <-time.After(5 * time.Minute):
		return nil, fmt.Errorf("authorization timeout")
	}
}

// authenticateClientCredentials performs the client credentials flow.
func (o *OAuth2Auth) authenticateClientCredentials(ctx context.Context) (*Token, error) {
	token, err := o.ccConfig.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return convertOAuth2Token(token), nil
}

// authenticatePassword performs the password credentials flow.
func (o *OAuth2Auth) authenticatePassword(ctx context.Context) (*Token, error) {
	token, err := o.oauth2Config.PasswordCredentialsToken(ctx, o.config.Username, o.config.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return convertOAuth2Token(token), nil
}

// authenticateDeviceCode performs the device authorization flow.
func (o *OAuth2Auth) authenticateDeviceCode(ctx context.Context) (*Token, error) {
	// Request device code
	deviceAuth, err := o.requestDeviceCode(ctx)
	if err != nil {
		return nil, err
	}

	// Display user code to user
	fmt.Printf("\nDevice Authorization:\n")
	fmt.Printf("User Code: %s\n", deviceAuth.UserCode)
	fmt.Printf("Verification URL: %s\n", deviceAuth.VerificationURL)
	if deviceAuth.VerificationURLComplete != "" {
		fmt.Printf("Or visit: %s\n", deviceAuth.VerificationURLComplete)
	}
	fmt.Printf("\nWaiting for authorization...\n")

	// Poll for token
	return o.pollDeviceToken(ctx, deviceAuth)
}

// startCallbackServer starts a local HTTP server to receive OAuth callbacks.
func (o *OAuth2Auth) startCallbackServer() (*http.Server, chan string, error) {
	callbackChan := make(chan string, 1)

	// Parse redirect URL to get port
	redirectURL := o.config.RedirectURL
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/callback"
	}

	u, err := url.Parse(redirectURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid redirect URL: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(u.Path, func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			callbackChan <- ""
			return
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprintf(w, "<html><body><h1>Authorization successful!</h1><p>You can close this window.</p></body></html>")
		callbackChan <- code
	})

	addr := u.Host
	if !strings.Contains(addr, ":") {
		addr = "localhost:8080"
	}

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	go func() { _ = server.Serve(listener) }()

	return server, callbackChan, nil
}

// buildAuthURL builds the authorization URL with PKCE if enabled.
func (o *OAuth2Auth) buildAuthURL() string {
	state := generateRandomString(32)

	var opts []oauth2.AuthCodeOption
	if o.config.PKCE {
		verifier := generateRandomString(64)
		o.pkceVerifier = verifier

		// Create code challenge
		h := sha256.New()
		h.Write([]byte(verifier))
		challenge := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", challenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	// Add any additional endpoint parameters
	for key, value := range o.config.EndpointParams {
		opts = append(opts, oauth2.SetAuthURLParam(key, value))
	}

	return o.oauth2Config.AuthCodeURL(state, opts...)
}

// DeviceAuthResponse represents the response from a device authorization request.
type DeviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURL         string `json:"verification_uri"`
	VerificationURLComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// requestDeviceCode requests a device code from the authorization server.
func (o *OAuth2Auth) requestDeviceCode(ctx context.Context) (*DeviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", o.config.ClientID)
	if len(o.config.Scopes) > 0 {
		data.Set("scope", strings.Join(o.config.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.DeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed: %s - %s", resp.Status, string(body))
	}

	var deviceAuth DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceAuth); err != nil {
		return nil, fmt.Errorf("failed to decode device auth response: %w", err)
	}

	if deviceAuth.Interval == 0 {
		deviceAuth.Interval = 5 // Default to 5 seconds
	}

	return &deviceAuth, nil
}

// pollDeviceToken polls the token endpoint until authorization is complete.
func (o *OAuth2Auth) pollDeviceToken(ctx context.Context, deviceAuth *DeviceAuthResponse) (*Token, error) {
	interval := time.Duration(deviceAuth.Interval) * time.Second
	expiresAt := time.Now().Add(time.Duration(deviceAuth.ExpiresIn) * time.Second)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authorization cancelled")
		case <-ticker.C:
			if time.Now().After(expiresAt) {
				return nil, fmt.Errorf("device code expired")
			}

			token, err := o.checkDeviceToken(ctx, deviceAuth.DeviceCode)
			if err != nil {
				// Check for specific errors
				if strings.Contains(err.Error(), "authorization_pending") {
					continue
				}
				if strings.Contains(err.Error(), "slow_down") {
					interval = interval * 2
					ticker.Reset(interval)
					continue
				}
				return nil, err
			}

			return token, nil
		}
	}
}

// checkDeviceToken attempts to exchange the device code for a token.
func (o *OAuth2Auth) checkDeviceToken(ctx context.Context, deviceCode string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)
	data.Set("client_id", o.config.ClientID)

	if o.config.ClientSecret != "" {
		data.Set("client_secret", o.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check device token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil {
			return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var oauth2Token oauth2.Token
	if err := json.Unmarshal(body, &oauth2Token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return convertOAuth2Token(&oauth2Token), nil
}

// convertOAuth2Token converts an oauth2.Token to our Token type.
func convertOAuth2Token(token *oauth2.Token) *Token {
	t := &Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}

	// Extract scopes if present
	if scope, ok := token.Extra("scope").(string); ok {
		t.Scopes = strings.Split(scope, " ")
	}

	// Store other extra fields
	t.Extra = make(map[string]interface{})
	// Note: oauth2.Token.Extra returns a map[string]interface{} but only accepts key for lookup
	// We can only extract known keys
	if idToken := token.Extra("id_token"); idToken != nil {
		t.Extra["id_token"] = idToken
	}

	return t
}

// convertEndpointParams converts string map to url.Values.
func convertEndpointParams(params map[string]string) url.Values {
	if params == nil {
		return nil
	}

	values := make(url.Values)
	for key, value := range params {
		values.Set(key, value)
	}
	return values
}

// generateRandomString generates a random string of the specified length.
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)[:length]
}
