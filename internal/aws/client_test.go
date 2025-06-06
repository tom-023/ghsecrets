package aws

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
	// Skip if AWS credentials are not configured
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	client, err := NewClient("us-east-1")
	
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, "us-east-1", client.region)
}

func TestNewClientDifferentRegions(t *testing.T) {
	// Skip if AWS credentials are not configured
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	regions := []string{"us-east-1", "us-west-2", "eu-west-1"}
	
	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			client, err := NewClient(region)
			require.NoError(t, err)
			assert.Equal(t, region, client.region)
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
			name:     "secret exists error - version",
			err:      fmt.Errorf("ResourceExistsException: The operation failed because the secret version already exists."),
			expected: true,
		},
		{
			name:     "secret exists error - resource",
			err:      fmt.Errorf("ResourceExistsException: A resource with the ID you requested already exists."),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("AccessDeniedException: You do not have permission to perform this action."),
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

// Integration test - requires valid AWS credentials
func TestCreateOrUpdateSecret(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping integration test: AWS credentials not configured")
	}

	client, err := NewClient("us-east-1")
	require.NoError(t, err)

	ctx := context.Background()
	secretName := "ghsecrets-test-" + time.Now().Format("20060102150405")
	secretValue := "test-value"
	description := "Test secret for ghsecrets"

	// Create secret
	err = client.CreateOrUpdateSecret(ctx, secretName, secretValue, description)
	require.NoError(t, err)

	// Update secret
	newValue := "updated-value"
	err = client.CreateOrUpdateSecret(ctx, secretName, newValue, description)
	require.NoError(t, err)

	// Verify
	value, err := client.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, newValue, value)

	// Cleanup
	err = client.DeleteSecret(ctx, secretName)
	require.NoError(t, err)
}