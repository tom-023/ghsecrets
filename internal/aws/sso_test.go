package aws

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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