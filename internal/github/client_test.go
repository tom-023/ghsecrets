package github

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token", "owner", "repo")

	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, "owner", client.owner)
	assert.Equal(t, "repo", client.repo)
	assert.Equal(t, "test-token", client.token)
}

func TestEncryptSecret(t *testing.T) {
	// This is a test public key (base64 encoded 32-byte key)
	publicKey := "RRjlhKlgU2SicuhpgO3vV8BDVmFpNMIYY0k8mp9FqrU="
	secret := "my-secret-value"

	encrypted, err := encryptSecret(publicKey, secret)

	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)
	assert.NotEqual(t, secret, encrypted)
}

func TestEncryptSecretInvalidKey(t *testing.T) {
	tests := []struct {
		name      string
		publicKey string
		secret    string
		wantErr   string
	}{
		{
			name:      "invalid base64",
			publicKey: "not-valid-base64!@#$",
			secret:    "my-secret",
			wantErr:   "failed to decode public key",
		},
		{
			name:      "empty key",
			publicKey: "",
			secret:    "my-secret",
			wantErr:   "failed to encrypt",
		},
		{
			name:      "wrong key length",
			publicKey: "c2hvcnQ=", // "short" in base64
			secret:    "my-secret",
			wantErr:   "failed to encrypt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptSecret(tt.publicKey, tt.secret)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestActionsPublicKey(t *testing.T) {
	pk := &ActionsPublicKey{
		KeyID: "123456",
		Key:   "RRjlhKlgU2SicuhpgO3vV8BDVmFpNMIYY0k8mp9FqrU=",
	}

	assert.Equal(t, "123456", pk.KeyID)
	assert.Equal(t, "RRjlhKlgU2SicuhpgO3vV8BDVmFpNMIYY0k8mp9FqrU=", pk.Key)
}

// Integration test - requires valid GitHub token and repository
func TestCreateOrUpdateSecret(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	token := os.Getenv("TEST_GITHUB_TOKEN")
	owner := os.Getenv("TEST_GITHUB_OWNER")
	repo := os.Getenv("TEST_GITHUB_REPO")

	if token == "" || owner == "" || repo == "" {
		t.Skip("Skipping integration test: TEST_GITHUB_TOKEN, TEST_GITHUB_OWNER, or TEST_GITHUB_REPO not set")
	}

	client := NewClient(token, owner, repo)
	ctx := context.Background()

	err := client.CreateOrUpdateSecret(ctx, "TEST_SECRET", "test-value-"+time.Now().Format("20060102150405"))
	assert.NoError(t, err)
}