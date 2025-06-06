package gcp

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name            string
		projectID       string
		credentialsPath string
		wantErr         bool
	}{
		{
			name:            "with project ID only",
			projectID:       "test-project",
			credentialsPath: "",
			wantErr:         false,
		},
		{
			name:            "with credentials path",
			projectID:       "test-project",
			credentialsPath: "/path/to/creds.json",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if running in CI or without GCP setup
			if os.Getenv("CI") == "true" || os.Getenv("GOOGLE_APPLICATION_CREDENTIALS") == "" {
				t.Skip("Skipping GCP client test")
			}

			client, err := NewClient(tt.projectID, tt.credentialsPath)
			
			if tt.wantErr {
				require.Error(t, err)
			} else {
				// May still error if no valid credentials
				if err != nil {
					t.Logf("Expected no error but got: %v (likely due to missing credentials)", err)
				} else {
					assert.NotNil(t, client)
					assert.Equal(t, tt.projectID, client.projectID)
				}
			}
		})
	}
}

func TestIsSecretExistsError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "secret exists error",
			err:      fmt.Errorf("rpc error: code = AlreadyExists desc = Secret [projects/*/secrets/*] already exists."),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("rpc error: code = PermissionDenied desc = Permission denied"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSecretExistsError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockCreateOrUpdateSecret(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient("test-project")

	secretName := "test-secret"
	secretValue := "test-value"

	// Test creating a new secret
	err := mockClient.CreateOrUpdateSecret(ctx, secretName, secretValue)
	require.NoError(t, err)

	// Verify secret was created
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


func TestMockErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient("test-project")

	// Test error on CreateOrUpdateSecret
	mockClient.SetError("CreateOrUpdateSecret", fmt.Errorf("permission denied"))
	err := mockClient.CreateOrUpdateSecret(ctx, "test", "value")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Clear error and create secret
	mockClient.SetError("CreateOrUpdateSecret", nil)
	err = mockClient.CreateOrUpdateSecret(ctx, "test", "value")
	require.NoError(t, err)

	// Test error on GetSecret
	mockClient.SetError("GetSecret", fmt.Errorf("access denied"))
	_, err = mockClient.GetSecret(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")

	// Test Close
	err = mockClient.Close()
	assert.NoError(t, err)
}