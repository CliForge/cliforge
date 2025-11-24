package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	if factory == nil {
		t.Fatal("NewFactory() returned nil")
	}
}

func TestFactory_Create(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.StorageConfig
		cliName string
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			cliName: "test-cli",
			wantErr: true,
		},
		{
			name: "memory storage",
			config: &types.StorageConfig{
				Type: types.StorageTypeMemory,
			},
			cliName: "test-cli",
			wantErr: false,
		},
		{
			name: "file storage",
			config: &types.StorageConfig{
				Type: types.StorageTypeFile,
				Path: "/tmp/test-token.json",
			},
			cliName: "test-cli",
			wantErr: false,
		},
		{
			name: "unsupported storage type",
			config: &types.StorageConfig{
				Type: types.StorageType("invalid"),
			},
			cliName: "test-cli",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory()
			storage, err := factory.Create(tt.config, tt.cliName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && storage == nil {
				t.Error("Create() returned nil storage")
			}
		})
	}
}

// Mock storage for testing MultiStorage
type mockStorage struct {
	saveFunc   func(ctx context.Context, token *types.Token) error
	loadFunc   func(ctx context.Context) (*types.Token, error)
	deleteFunc func(ctx context.Context) error
}

func (m *mockStorage) SaveToken(ctx context.Context, token *types.Token) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, token)
	}
	return nil
}

func (m *mockStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return nil, fmt.Errorf("no token")
}

func (m *mockStorage) DeleteToken(ctx context.Context) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx)
	}
	return nil
}

func TestNewMultiStorage(t *testing.T) {
	stor1 := NewMemoryStorage()
	stor2 := NewMemoryStorage()

	multi := NewMultiStorage(stor1, stor2)
	if multi == nil {
		t.Fatal("NewMultiStorage() returned nil")
	}

	if len(multi.storages) != 2 {
		t.Errorf("NewMultiStorage() storages count = %d, want 2", len(multi.storages))
	}
}

func TestMultiStorage_SaveToken(t *testing.T) {
	tests := []struct {
		name     string
		storages []TokenStorage
		wantErr  bool
	}{
		{
			name: "save to all storages successfully",
			storages: []TokenStorage{
				NewMemoryStorage(),
				NewMemoryStorage(),
			},
			wantErr: false,
		},
		{
			name: "partial success",
			storages: []TokenStorage{
				NewMemoryStorage(),
				&mockStorage{
					saveFunc: func(ctx context.Context, token *types.Token) error {
						return fmt.Errorf("save failed")
					},
				},
			},
			wantErr: false, // Should succeed if at least one storage saves
		},
		{
			name: "all fail",
			storages: []TokenStorage{
				&mockStorage{
					saveFunc: func(ctx context.Context, token *types.Token) error {
						return fmt.Errorf("save failed 1")
					},
				},
				&mockStorage{
					saveFunc: func(ctx context.Context, token *types.Token) error {
						return fmt.Errorf("save failed 2")
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMultiStorage(tt.storages...)
			ctx := context.Background()
			token := &types.Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(time.Hour),
			}

			err := multi.SaveToken(ctx, token)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMultiStorage_LoadToken(t *testing.T) {
	ctx := context.Background()
	testToken := &types.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}

	tests := []struct {
		name     string
		storages []TokenStorage
		wantErr  bool
		wantNil  bool
	}{
		{
			name: "load from first storage",
			storages: []TokenStorage{
				&mockStorage{
					loadFunc: func(ctx context.Context) (*types.Token, error) {
						return testToken, nil
					},
				},
				NewMemoryStorage(),
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "load from second storage when first fails",
			storages: []TokenStorage{
				&mockStorage{
					loadFunc: func(ctx context.Context) (*types.Token, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockStorage{
					loadFunc: func(ctx context.Context) (*types.Token, error) {
						return testToken, nil
					},
				},
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name: "all fail",
			storages: []TokenStorage{
				&mockStorage{
					loadFunc: func(ctx context.Context) (*types.Token, error) {
						return nil, fmt.Errorf("not found")
					},
				},
				&mockStorage{
					loadFunc: func(ctx context.Context) (*types.Token, error) {
						return nil, fmt.Errorf("not found")
					},
				},
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMultiStorage(tt.storages...)
			token, err := multi.LoadToken(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantNil && token != nil {
				t.Error("LoadToken() returned non-nil token when error expected")
			}
			if !tt.wantNil && !tt.wantErr && token == nil {
				t.Error("LoadToken() returned nil token")
			}
		})
	}
}

func TestMultiStorage_DeleteToken(t *testing.T) {
	tests := []struct {
		name     string
		storages []TokenStorage
		wantErr  bool
	}{
		{
			name: "delete from all storages successfully",
			storages: []TokenStorage{
				NewMemoryStorage(),
				NewMemoryStorage(),
			},
			wantErr: false,
		},
		{
			name: "partial failure",
			storages: []TokenStorage{
				NewMemoryStorage(),
				&mockStorage{
					deleteFunc: func(ctx context.Context) error {
						return fmt.Errorf("delete failed")
					},
				},
			},
			wantErr: true, // Returns last error
		},
		{
			name: "all fail",
			storages: []TokenStorage{
				&mockStorage{
					deleteFunc: func(ctx context.Context) error {
						return fmt.Errorf("delete failed 1")
					},
				},
				&mockStorage{
					deleteFunc: func(ctx context.Context) error {
						return fmt.Errorf("delete failed 2")
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			multi := NewMultiStorage(tt.storages...)
			ctx := context.Background()

			err := multi.DeleteToken(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
