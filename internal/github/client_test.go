package github

import (
	"context"
	"fmt"
	"testing"

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

func TestMockCreateOrUpdateSecret(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient("test-token", "test-owner", "test-repo")

	secretName := "TEST_SECRET"
	secretValue := "test-value"

	// Test creating a secret
	err := mockClient.CreateOrUpdateSecret(ctx, secretName, secretValue)
	require.NoError(t, err)

	// Verify secret was created (using our mock GetSecret method)
	value, err := mockClient.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, secretValue, value)

	// Test updating existing secret
	newValue := "updated-value"
	err = mockClient.CreateOrUpdateSecret(ctx, secretName, newValue)
	require.NoError(t, err)

	// Verify secret was updated
	value, err = mockClient.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, newValue, value)
}

func TestMockAuthenticationError(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient("", "test-owner", "test-repo")

	// Test with no token
	err := mockClient.CreateOrUpdateSecret(ctx, "TEST_SECRET", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestMockErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient("test-token", "test-owner", "test-repo")

	// Test error on CreateOrUpdateSecret
	mockClient.SetError("CreateOrUpdateSecret", fmt.Errorf("API rate limit exceeded"))
	err := mockClient.CreateOrUpdateSecret(ctx, "test", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API rate limit exceeded")

	// Clear error and create secret
	mockClient.SetError("CreateOrUpdateSecret", nil)
	err = mockClient.CreateOrUpdateSecret(ctx, "test", "value")
	require.NoError(t, err)

	// Test error on GetSecret
	mockClient.SetError("GetSecret", fmt.Errorf("not found"))
	_, err = mockClient.GetSecret(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}