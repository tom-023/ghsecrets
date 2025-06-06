package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetGitHubToken(t *testing.T) {
	tests := []struct {
		name        string
		configToken string
		envToken    string
		setupEnv    func()
		cleanupEnv  func()
		wantToken   string
		wantErr     bool
	}{
		{
			name:        "config token takes priority",
			configToken: "config-token",
			envToken:    "env-token",
			setupEnv:    func() { os.Setenv("GITHUB_TOKEN", "env-token") },
			cleanupEnv:  func() { os.Unsetenv("GITHUB_TOKEN") },
			wantToken:   "config-token",
			wantErr:     false,
		},
		{
			name:        "env token when no config token",
			configToken: "",
			envToken:    "env-token",
			setupEnv:    func() { os.Setenv("GITHUB_TOKEN", "env-token") },
			cleanupEnv:  func() { os.Unsetenv("GITHUB_TOKEN") },
			wantToken:   "env-token",
			wantErr:     false,
		},
		{
			name:        "no token available",
			configToken: "",
			envToken:    "",
			setupEnv:    func() { os.Unsetenv("GITHUB_TOKEN") },
			cleanupEnv:  func() {},
			wantToken:   "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env var
			originalToken := os.Getenv("GITHUB_TOKEN")
			defer os.Setenv("GITHUB_TOKEN", originalToken)

			// Setup
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			// Test
			token, err := GetGitHubToken(tt.configToken)

			// Assert
			if tt.wantErr {
				if err == nil {
					t.Logf("Expected error but got token: %s", token)
					// Check if gh CLI provided a token
					if token != "" {
						t.Skip("gh CLI is authenticated, skipping negative test")
					}
				}
				require.Error(t, err)
				assert.Contains(t, err.Error(), "GitHub token not found")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestGetGHCLIToken(t *testing.T) {
	// This test will fail if gh CLI is not installed or not authenticated
	// We'll skip it in CI environments
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping gh CLI test in CI environment")
	}

	token, err := getGHCLIToken()
	if err != nil {
		// If gh is not authenticated, we expect an error
		assert.Contains(t, err.Error(), "gh CLI not authenticated")
	} else {
		// If authenticated, we should get a non-empty token
		assert.NotEmpty(t, token)
	}
}