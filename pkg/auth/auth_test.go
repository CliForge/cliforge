package auth

import (
	"testing"
	"time"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		token     *Token
		wantExpired bool
	}{
		{
			name: "valid token not expired",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "token expiring in 15 seconds (within buffer)",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(15 * time.Second),
			},
			wantExpired: true,
		},
		{
			name: "token with zero expiry",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Time{},
			},
			wantExpired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.wantExpired {
				t.Errorf("Token.IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

func TestToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token *Token
		want  bool
	}{
		{
			name: "valid token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "empty access token",
			token: &Token{
				AccessToken: "",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			want: false,
		},
		{
			name: "valid token no expiry",
			token: &Token{
				AccessToken: "test-token",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("Token.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthType(t *testing.T) {
	types := []AuthType{
		AuthTypeAPIKey,
		AuthTypeOAuth2,
		AuthTypeBasic,
		AuthTypeNone,
	}

	expected := []string{"apikey", "oauth2", "basic", "none"}

	for i, authType := range types {
		if string(authType) != expected[i] {
			t.Errorf("AuthType[%d] = %v, want %v", i, authType, expected[i])
		}
	}
}

func TestAPIKeyLocation(t *testing.T) {
	locations := []APIKeyLocation{
		APIKeyLocationHeader,
		APIKeyLocationQuery,
	}

	expected := []string{"header", "query"}

	for i, loc := range locations {
		if string(loc) != expected[i] {
			t.Errorf("APIKeyLocation[%d] = %v, want %v", i, loc, expected[i])
		}
	}
}

func TestOAuth2Flow(t *testing.T) {
	flows := []OAuth2Flow{
		OAuth2FlowAuthorizationCode,
		OAuth2FlowClientCredentials,
		OAuth2FlowPassword,
		OAuth2FlowDeviceCode,
	}

	expected := []string{
		"authorization_code",
		"client_credentials",
		"password",
		"device_code",
	}

	for i, flow := range flows {
		if string(flow) != expected[i] {
			t.Errorf("OAuth2Flow[%d] = %v, want %v", i, flow, expected[i])
		}
	}
}

func TestStorageType(t *testing.T) {
	types := []StorageType{
		StorageTypeFile,
		StorageTypeKeyring,
		StorageTypeMemory,
	}

	expected := []string{"file", "keyring", "memory"}

	for i, storType := range types {
		if string(storType) != expected[i] {
			t.Errorf("StorageType[%d] = %v, want %v", i, storType, expected[i])
		}
	}
}
