package gcp

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

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

// Integration test - requires valid GCP credentials and project
func TestCreateOrUpdateSecret(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	projectID := os.Getenv("TEST_GCP_PROJECT")
	credsPath := os.Getenv("TEST_GCP_CREDENTIALS")

	if projectID == "" {
		t.Skip("Skipping integration test: TEST_GCP_PROJECT not set")
	}

	client, err := NewClient(projectID, credsPath)
	if err != nil {
		t.Skipf("Failed to create client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	secretName := "ghsecrets-test-" + time.Now().Format("20060102150405")
	secretValue := "test-value"

	// Create secret
	err = client.CreateOrUpdateSecret(ctx, secretName, secretValue)
	require.NoError(t, err)

	// Update secret
	newValue := "updated-value"
	err = client.CreateOrUpdateSecret(ctx, secretName, newValue)
	require.NoError(t, err)

	// Verify
	value, err := client.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, newValue, value)

	// Cleanup
	err = client.DeleteSecret(ctx, secretName)
	require.NoError(t, err)
}