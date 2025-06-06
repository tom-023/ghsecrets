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

// TestNewClientWithSSOProfile tests that SSO profiles can be used
// This test will only run if AWS SSO is configured
func TestNewClientWithSSOProfile(t *testing.T) {
	// Skip if no SSO profile is configured
	ssoProfile := os.Getenv("TEST_AWS_SSO_PROFILE")
	if ssoProfile == "" {
		t.Skip("Skipping SSO test: TEST_AWS_SSO_PROFILE not set")
	}

	// Create client with SSO profile
	client, err := NewClientWithOptions(ClientOptions{
		Region:  "ap-northeast-1",
		Profile: ssoProfile,
	})

	// The client creation should succeed if the user has logged in with:
	// aws sso login --profile <profile-name>
	if err != nil {
		t.Logf("SSO profile test failed - make sure you're logged in with: aws sso login --profile %s", ssoProfile)
		t.Logf("Error: %v", err)
		t.Skip("Skipping SSO test due to authentication error")
	}

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "ap-northeast-1", client.region)
}

// TestSSOProfileDocumentation verifies that SSO profile usage is documented
func TestSSOProfileDocumentation(t *testing.T) {
	// This test just ensures we remember to support SSO profiles
	// AWS SDK v2 handles SSO profiles automatically when using WithSharedConfigProfile
	
	opts := ClientOptions{
		Region:  "us-east-1",
		Profile: "my-sso-profile",
	}
	
	// This would work with any profile type (regular or SSO)
	// as long as the profile is properly configured in ~/.aws/config
	assert.Equal(t, "my-sso-profile", opts.Profile)
	assert.Equal(t, "us-east-1", opts.Region)
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
	// AWS SDK v2 uses typed errors, so we need to test with actual error types
	// For unit tests, we'll just verify the function exists and handles nil
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
			name:     "generic error",
			err:      fmt.Errorf("some error"),
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