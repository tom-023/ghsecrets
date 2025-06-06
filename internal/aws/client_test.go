package aws

import (
	"context"
	"fmt"
	"os"
	"testing"

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

func TestNewClientWithOptions(t *testing.T) {
	// Skip if AWS credentials are not configured
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		t.Skip("Skipping test: AWS credentials not configured")
	}

	tests := []struct {
		name    string
		opts    ClientOptions
		wantErr bool
	}{
		{
			name: "with region only",
			opts: ClientOptions{
				Region: "us-west-2",
			},
			wantErr: false,
		},
		{
			name: "with region and profile",
			opts: ClientOptions{
				Region:  "eu-west-1",
				Profile: "default",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClientWithOptions(tt.opts)
			
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.opts.Region, client.region)
			}
		})
	}
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

func TestMockCreateOrUpdateSecret(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()

	secretName := "test-secret"
	secretValue := "test-value"
	description := "Test secret"

	// Test creating a new secret
	err := mockClient.CreateOrUpdateSecret(ctx, secretName, secretValue, description)
	require.NoError(t, err)

	// Verify secret was created
	value, err := mockClient.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, secretValue, value)

	// Test updating existing secret
	newValue := "updated-value"
	err = mockClient.CreateOrUpdateSecret(ctx, secretName, newValue, description)
	require.NoError(t, err)

	// Verify secret was updated
	value, err = mockClient.GetSecret(ctx, secretName)
	require.NoError(t, err)
	assert.Equal(t, newValue, value)
}


func TestMockErrorHandling(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()

	// Test error on CreateOrUpdateSecret
	mockClient.SetError("CreateOrUpdateSecret", fmt.Errorf("permission denied"))
	err := mockClient.CreateOrUpdateSecret(ctx, "test", "value", "desc")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Clear error and create secret
	mockClient.SetError("CreateOrUpdateSecret", nil)
	err = mockClient.CreateOrUpdateSecret(ctx, "test", "value", "desc")
	require.NoError(t, err)

	// Test error on GetSecret
	mockClient.SetError("GetSecret", fmt.Errorf("access denied"))
	_, err = mockClient.GetSecret(ctx, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}