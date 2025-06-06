package github

import (
	"context"
	"fmt"
	"sync"
)

// MockClient is a mock implementation of GitHub client for testing
type MockClient struct {
	mu      sync.Mutex
	secrets map[string]string
	errors  map[string]error
	owner   string
	repo    string
	token   string
}

// NewMockClient creates a new mock GitHub client
func NewMockClient(token, owner, repo string) *MockClient {
	return &MockClient{
		owner:   owner,
		repo:    repo,
		token:   token,
		secrets: make(map[string]string),
		errors:  make(map[string]error),
	}
}

// SetError sets an error to be returned for a specific operation
func (m *MockClient) SetError(operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[operation] = err
}

// CreateOrUpdateSecret mocks the CreateOrUpdateSecret method
func (m *MockClient) CreateOrUpdateSecret(ctx context.Context, name, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["CreateOrUpdateSecret"]; err != nil {
		return err
	}

	// Simulate token validation
	if m.token == "" {
		return fmt.Errorf("authentication required")
	}

	// Store the encrypted value (in real implementation this would be encrypted)
	m.secrets[name] = value
	return nil
}

// GetSecret mocks getting a secret (note: GitHub API doesn't support this)
func (m *MockClient) GetSecret(ctx context.Context, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["GetSecret"]; err != nil {
		return "", err
	}

	value, exists := m.secrets[name]
	if !exists {
		return "", fmt.Errorf("secret not found: %s", name)
	}

	return value, nil
}

