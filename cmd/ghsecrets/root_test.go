package ghsecrets

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	// Test help output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	
	err := rootCmd.Execute()
	require.NoError(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "ghsecrets is a CLI tool")
	assert.Contains(t, output, "Available Commands:")
}

func TestInitConfig(t *testing.T) {
	// Create temp directory for test config
	tmpDir, err := os.MkdirTemp("", "ghsecrets-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test config file
	configPath := filepath.Join(tmpDir, ".ghsecrets.yaml")
	configContent := `
github:
  owner: test-owner
  repo: test-repo
aws:
  region: us-west-2
gcp:
  project: test-project
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Reset viper for clean test
	viper.Reset()

	// Set config file
	cfgFile = configPath
	initConfig()

	// Verify config was loaded
	assert.Equal(t, "test-owner", viper.GetString("github.owner"))
	assert.Equal(t, "test-repo", viper.GetString("github.repo"))
	assert.Equal(t, "us-west-2", viper.GetString("aws.region"))
	assert.Equal(t, "test-project", viper.GetString("gcp.project"))
}

func TestExecuteWithInvalidCommand(t *testing.T) {
	// Create a new root command for testing
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	// Set invalid args
	os.Args = []string{"ghsecrets", "invalid-command"}
	
	// The Execute function calls os.Exit(1) on error, 
	// so we'll test the command directly instead
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"invalid-command"})
	
	err := rootCmd.Execute()
	require.Error(t, err)
	
	output := buf.String()
	assert.Contains(t, output, "Error:")
}