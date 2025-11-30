package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents parsed JWT claims
type JWTClaims struct {
	Type              string
	Subject           string
	PreferredUsername string
	Username          string
	Email             string
	ExpiresAt         time.Time
	IssuedAt          time.Time
	Issuer            string
}

// TokenType represents the type of token
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
	TokenTypeOffline TokenType = "offline"
	TokenTypeBearer  TokenType = "bearer"
	TokenTypeUnknown TokenType = "unknown"
)

// ParseJWT parses JWT WITHOUT validation (for claim inspection only)
// Returns error only for malformed tokens, NOT for expired/invalid signatures
func ParseJWT(tokenString string) (*JWTClaims, error) {
	// Parse without validation to inspect claims only
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from token")
	}

	jwtClaims := &JWTClaims{}

	// Extract typ claim
	if typ, ok := claims["typ"].(string); ok {
		jwtClaims.Type = typ
	}

	// Extract sub claim
	if sub, ok := claims["sub"].(string); ok {
		jwtClaims.Subject = sub
	}

	// Extract preferred_username claim
	if preferredUsername, ok := claims["preferred_username"].(string); ok {
		jwtClaims.PreferredUsername = preferredUsername
	}

	// Extract username claim
	if username, ok := claims["username"].(string); ok {
		jwtClaims.Username = username
	}

	// Extract email claim
	if email, ok := claims["email"].(string); ok {
		jwtClaims.Email = email
	}

	// Extract exp claim
	if exp, ok := claims["exp"].(float64); ok {
		jwtClaims.ExpiresAt = time.Unix(int64(exp), 0)
	}

	// Extract iat claim
	if iat, ok := claims["iat"].(float64); ok {
		jwtClaims.IssuedAt = time.Unix(int64(iat), 0)
	}

	// Extract iss claim
	if iss, ok := claims["iss"].(string); ok {
		jwtClaims.Issuer = iss
	}

	return jwtClaims, nil
}

// DetectTokenType determines token type from "typ" claim
// Logic from ROSA (hack/rosa-cli/rosa/cmd/login/cmd.go lines 397-423):
// - "Bearer" or "" → access token
// - "Refresh" or "Offline" → refresh/offline token
func DetectTokenType(tokenString string) (TokenType, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return TokenTypeUnknown, err
	}

	typ := claims.Type

	// Normalize type string for comparison
	typLower := strings.ToLower(typ)

	switch typLower {
	case "bearer":
		return TokenTypeBearer, nil
	case "":
		return TokenTypeAccess, nil
	case "refresh":
		return TokenTypeRefresh, nil
	case "offline":
		return TokenTypeOffline, nil
	default:
		return TokenTypeUnknown, nil
	}
}

// JWETokenHeader represents JWE token header structure
type JWETokenHeader struct {
	Algorithm   string `json:"alg"`
	Encryption  string `json:"enc"`
	ContentType string `json:"cty,omitempty"`
}

// IsEncryptedToken checks if token is JWE (5 parts) vs JWT (3 parts)
// ROSA: hack/rosa-cli/rosa/pkg/config/token.go lines 35-54
func IsEncryptedToken(tokenString string) bool {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 5 {
		return false
	}

	// Decode and parse JWE header to verify it's actually encrypted
	encoded := fmt.Sprintf("%s==", parts[0])
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(decoded) == 0 {
		return false
	}

	header := new(JWETokenHeader)
	err = json.Unmarshal(decoded, header)
	if err != nil {
		return false
	}

	// Verify it has encryption and content type is JWT
	if header.Encryption != "" && header.ContentType == "JWT" {
		return true
	}

	return false
}

// ExtractUsername gets username from token claims
// Tries "preferred_username" first, then "username"
// ROSA: hack/rosa-cli/rosa/cmd/login/cmd.go lines 300-304
func ExtractUsername(tokenString string) (string, error) {
	claims, err := ParseJWT(tokenString)
	if err != nil {
		return "", err
	}

	// Try preferred_username first
	if claims.PreferredUsername != "" {
		return claims.PreferredUsername, nil
	}

	// Fallback to username
	if claims.Username != "" {
		return claims.Username, nil
	}

	// No username found
	return "", fmt.Errorf("no username found in token claims")
}

// parseJWTHeader parses the JWT header without validation
func parseJWTHeader(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to decode header: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	return header, nil
}
