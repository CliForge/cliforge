package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create test JWT tokens
func createTestToken(header, claims map[string]interface{}) string {
	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create a fake signature (we're not validating)
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	return headerEncoded + "." + claimsEncoded + "." + signature
}

func TestParseJWT_ValidAccessToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	claims := map[string]interface{}{
		"typ":                "Bearer",
		"sub":                "user-123",
		"preferred_username": "testuser",
		"email":              "test@example.com",
		"exp":                float64(time.Now().Add(1 * time.Hour).Unix()),
		"iat":                float64(time.Now().Unix()),
		"iss":                "https://sso.redhat.com/auth/realms/redhat-external",
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.Equal(t, "Bearer", result.Type)
	assert.Equal(t, "user-123", result.Subject)
	assert.Equal(t, "testuser", result.PreferredUsername)
	assert.Equal(t, "test@example.com", result.Email)
	assert.False(t, result.ExpiresAt.IsZero())
	assert.False(t, result.IssuedAt.IsZero())
	assert.Equal(t, "https://sso.redhat.com/auth/realms/redhat-external", result.Issuer)
}

func TestParseJWT_ValidRefreshToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := map[string]interface{}{
		"typ":                "Refresh",
		"sub":                "user-456",
		"preferred_username": "refreshuser",
		"exp":                float64(time.Now().Add(24 * time.Hour).Unix()),
		"iat":                float64(time.Now().Unix()),
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.Equal(t, "Refresh", result.Type)
	assert.Equal(t, "user-456", result.Subject)
	assert.Equal(t, "refreshuser", result.PreferredUsername)
}

func TestParseJWT_ValidOfflineToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	claims := map[string]interface{}{
		"typ":      "Offline",
		"sub":      "user-789",
		"username": "offlineuser",
		"exp":      float64(time.Now().Add(30 * 24 * time.Hour).Unix()),
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.Equal(t, "Offline", result.Type)
	assert.Equal(t, "user-789", result.Subject)
	assert.Equal(t, "offlineuser", result.Username)
}

func TestParseJWT_ExpiredToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	claims := map[string]interface{}{
		"typ":                "Bearer",
		"sub":                "user-123",
		"preferred_username": "testuser",
		"exp":                float64(time.Now().Add(-1 * time.Hour).Unix()), // Expired 1 hour ago
		"iat":                float64(time.Now().Add(-2 * time.Hour).Unix()),
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	// Should still parse successfully even though expired
	require.NoError(t, err)
	assert.Equal(t, "Bearer", result.Type)
	assert.Equal(t, "user-123", result.Subject)
	assert.True(t, result.ExpiresAt.Before(time.Now())) // Verify it's expired
}

func TestParseJWT_InvalidBase64(t *testing.T) {
	invalidToken := "invalid-base64-header.invalid-base64-claims.invalid-signature"
	_, err := ParseJWT(invalidToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse JWT")
}

func TestParseJWT_InvalidJSON(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	// Create valid header but invalid claims JSON
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Invalid JSON in claims
	claimsEncoded := base64.RawURLEncoding.EncodeToString([]byte("{invalid json}"))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	token := headerEncoded + "." + claimsEncoded + "." + signature
	_, err := ParseJWT(token)

	assert.Error(t, err)
}

func TestParseJWT_MissingStandardClaims(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}
	// Claims with only minimal data
	claims := map[string]interface{}{
		"sub": "user-123",
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.Equal(t, "user-123", result.Subject)
	// All other fields should be empty/zero
	assert.Empty(t, result.Type)
	assert.Empty(t, result.PreferredUsername)
	assert.Empty(t, result.Username)
	assert.Empty(t, result.Email)
	assert.True(t, result.ExpiresAt.IsZero())
	assert.True(t, result.IssuedAt.IsZero())
	assert.Empty(t, result.Issuer)
}

func TestDetectTokenType_Bearer(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"typ": "Bearer", "sub": "user-123"}

	token := createTestToken(header, claims)
	tokenType, err := DetectTokenType(token)

	require.NoError(t, err)
	assert.Equal(t, TokenTypeBearer, tokenType)
}

func TestDetectTokenType_EmptyType(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"sub": "user-123"}

	token := createTestToken(header, claims)
	tokenType, err := DetectTokenType(token)

	require.NoError(t, err)
	assert.Equal(t, TokenTypeAccess, tokenType)
}

func TestDetectTokenType_Refresh(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"typ": "Refresh", "sub": "user-123"}

	token := createTestToken(header, claims)
	tokenType, err := DetectTokenType(token)

	require.NoError(t, err)
	assert.Equal(t, TokenTypeRefresh, tokenType)
}

func TestDetectTokenType_Offline(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"typ": "Offline", "sub": "user-123"}

	token := createTestToken(header, claims)
	tokenType, err := DetectTokenType(token)

	require.NoError(t, err)
	assert.Equal(t, TokenTypeOffline, tokenType)
}

func TestDetectTokenType_Unknown(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"typ": "CustomType", "sub": "user-123"}

	token := createTestToken(header, claims)
	tokenType, err := DetectTokenType(token)

	require.NoError(t, err)
	assert.Equal(t, TokenTypeUnknown, tokenType)
}

func TestDetectTokenType_InvalidToken(t *testing.T) {
	invalidToken := "not.a.valid.token"
	tokenType, err := DetectTokenType(invalidToken)

	assert.Error(t, err)
	assert.Equal(t, TokenTypeUnknown, tokenType)
}

func TestDetectTokenType_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		typ      string
		expected TokenType
	}{
		{"lowercase bearer", "bearer", TokenTypeBearer},
		{"uppercase bearer", "BEARER", TokenTypeBearer},
		{"mixed case bearer", "BeArEr", TokenTypeBearer},
		{"lowercase refresh", "refresh", TokenTypeRefresh},
		{"uppercase refresh", "REFRESH", TokenTypeRefresh},
		{"lowercase offline", "offline", TokenTypeOffline},
		{"uppercase offline", "OFFLINE", TokenTypeOffline},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
			claims := map[string]interface{}{"typ": tc.typ, "sub": "user-123"}

			token := createTestToken(header, claims)
			tokenType, err := DetectTokenType(token)

			require.NoError(t, err)
			assert.Equal(t, tc.expected, tokenType)
		})
	}
}

func TestIsEncryptedToken_JWE(t *testing.T) {
	// Create a valid JWE header (JWE uses URL-safe base64 without padding)
	jweHeader := map[string]interface{}{
		"alg": "RSA-OAEP",
		"enc": "A256GCM",
		"cty": "JWT",
	}
	jweHeaderJSON, _ := json.Marshal(jweHeader)
	jweHeaderEncoded := base64.RawURLEncoding.EncodeToString(jweHeaderJSON)

	// JWE has 5 parts: header.encrypted-key.iv.ciphertext.tag
	jweToken := jweHeaderEncoded + ".encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.True(t, result)
}

func TestIsEncryptedToken_JWE_MissingContentType(t *testing.T) {
	// JWE header without "cty" field
	jweHeader := map[string]interface{}{
		"alg": "RSA-OAEP",
		"enc": "A256GCM",
	}
	jweHeaderJSON, _ := json.Marshal(jweHeader)
	jweHeaderEncoded := base64.RawURLEncoding.EncodeToString(jweHeaderJSON)

	jweToken := jweHeaderEncoded + ".encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.False(t, result) // Should fail without cty="JWT"
}

func TestIsEncryptedToken_JWE_MissingEncryption(t *testing.T) {
	// JWE header without "enc" field
	jweHeader := map[string]interface{}{
		"alg": "RSA-OAEP",
		"cty": "JWT",
	}
	jweHeaderJSON, _ := json.Marshal(jweHeader)
	jweHeaderEncoded := base64.RawURLEncoding.EncodeToString(jweHeaderJSON)

	jweToken := jweHeaderEncoded + ".encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.False(t, result) // Should fail without enc field
}

func TestIsEncryptedToken_JWE_WrongContentType(t *testing.T) {
	// JWE header with wrong content type
	jweHeader := map[string]interface{}{
		"alg": "RSA-OAEP",
		"enc": "A256GCM",
		"cty": "XML", // Wrong content type
	}
	jweHeaderJSON, _ := json.Marshal(jweHeader)
	jweHeaderEncoded := base64.RawURLEncoding.EncodeToString(jweHeaderJSON)

	jweToken := jweHeaderEncoded + ".encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.False(t, result)
}

func TestIsEncryptedToken_JWT(t *testing.T) {
	// JWT has 3 parts
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{"sub": "user-123"}
	jwtToken := createTestToken(header, claims)

	result := IsEncryptedToken(jwtToken)

	assert.False(t, result)
}

func TestIsEncryptedToken_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{"single part", "token"},
		{"two parts", "header.payload"},
		{"four parts", "a.b.c.d"},
		{"six parts", "a.b.c.d.e.f"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsEncryptedToken(tc.token)
			// Should return false for anything that's not exactly 5 parts
			assert.False(t, result)
		})
	}
}

func TestIsEncryptedToken_InvalidBase64Header(t *testing.T) {
	// 5 parts but invalid base64 header
	jweToken := "!!!invalid-base64!!!.encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.False(t, result)
}

func TestIsEncryptedToken_InvalidJSONHeader(t *testing.T) {
	// 5 parts with valid base64 but invalid JSON
	invalidJSON := base64.RawURLEncoding.EncodeToString([]byte("{invalid json}"))
	jweToken := invalidJSON + ".encrypted-key.iv.ciphertext.tag"
	result := IsEncryptedToken(jweToken)

	assert.False(t, result)
}

func TestExtractUsername_PreferredUsername(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{
		"preferred_username": "preferreduser",
		"username":           "regularuser",
		"sub":                "user-123",
	}

	token := createTestToken(header, claims)
	username, err := ExtractUsername(token)

	require.NoError(t, err)
	assert.Equal(t, "preferreduser", username)
}

func TestExtractUsername_FallbackToUsername(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{
		"username": "regularuser",
		"sub":      "user-123",
	}

	token := createTestToken(header, claims)
	username, err := ExtractUsername(token)

	require.NoError(t, err)
	assert.Equal(t, "regularuser", username)
}

func TestExtractUsername_NeitherClaim(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{
		"sub": "user-123",
	}

	token := createTestToken(header, claims)
	username, err := ExtractUsername(token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no username found")
	assert.Empty(t, username)
}

func TestExtractUsername_InvalidToken(t *testing.T) {
	invalidToken := "not.a.valid.token"
	username, err := ExtractUsername(invalidToken)

	assert.Error(t, err)
	assert.Empty(t, username)
}

func TestParseJWT_AllClaimsEmpty(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{} // Completely empty claims

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.NotNil(t, result)
	// All fields should be empty/zero
	assert.Empty(t, result.Type)
	assert.Empty(t, result.Subject)
	assert.Empty(t, result.PreferredUsername)
	assert.Empty(t, result.Username)
	assert.Empty(t, result.Email)
	assert.True(t, result.ExpiresAt.IsZero())
	assert.True(t, result.IssuedAt.IsZero())
	assert.Empty(t, result.Issuer)
}

func TestParseJWT_WrongClaimTypes(t *testing.T) {
	header := map[string]interface{}{"alg": "RS256", "typ": "JWT"}
	claims := map[string]interface{}{
		"typ":                123,     // Should be string
		"sub":                true,    // Should be string
		"preferred_username": 456.789, // Should be string
		"exp":                "not-a-number",
		"iat":                "also-not-a-number",
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	// Type assertions should fail gracefully, leaving fields empty
	assert.Empty(t, result.Type)
	assert.Empty(t, result.Subject)
	assert.Empty(t, result.PreferredUsername)
	assert.True(t, result.ExpiresAt.IsZero())
	assert.True(t, result.IssuedAt.IsZero())
}

func TestParseJWT_RealWorldLikeToken(t *testing.T) {
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": "key-id-123",
	}
	now := time.Now()
	claims := map[string]interface{}{
		"typ":                "Bearer",
		"sub":                "f:abc123:john.doe",
		"preferred_username": "jdoe",
		"username":           "john.doe",
		"email":              "john.doe@redhat.com",
		"email_verified":     true,
		"name":               "John Doe",
		"given_name":         "John",
		"family_name":        "Doe",
		"exp":                float64(now.Add(5 * time.Minute).Unix()),
		"iat":                float64(now.Unix()),
		"auth_time":          float64(now.Unix()),
		"iss":                "https://sso.redhat.com/auth/realms/redhat-external",
		"aud":                []interface{}{"cloud-services", "account"},
		"azp":                "rosa-cli",
		"scope":              "openid email profile",
		"acr":                "1",
		"allowed-origins":    []interface{}{"https://console.redhat.com"},
		"realm_access": map[string]interface{}{
			"roles": []interface{}{"offline_access", "uma_authorization"},
		},
	}

	token := createTestToken(header, claims)
	result, err := ParseJWT(token)

	require.NoError(t, err)
	assert.Equal(t, "Bearer", result.Type)
	assert.Equal(t, "f:abc123:john.doe", result.Subject)
	assert.Equal(t, "jdoe", result.PreferredUsername)
	assert.Equal(t, "john.doe", result.Username)
	assert.Equal(t, "john.doe@redhat.com", result.Email)
	assert.Equal(t, "https://sso.redhat.com/auth/realms/redhat-external", result.Issuer)
	assert.WithinDuration(t, now.Add(5*time.Minute), result.ExpiresAt, time.Second)
	assert.WithinDuration(t, now, result.IssuedAt, time.Second)
}

func TestTokenType_StringValues(t *testing.T) {
	// Verify TokenType constants have correct string values
	assert.Equal(t, "access", string(TokenTypeAccess))
	assert.Equal(t, "refresh", string(TokenTypeRefresh))
	assert.Equal(t, "offline", string(TokenTypeOffline))
	assert.Equal(t, "bearer", string(TokenTypeBearer))
	assert.Equal(t, "unknown", string(TokenTypeUnknown))
}
