package gcp

import (
	"context"
	"fmt"
	"sync"
)

// MockClient is a mock implementation of GCP Secret Manager client for testing
type MockClient struct {
	mu        sync.Mutex
	secrets   map[string]string
	errors    map[string]error
	projectID string
}

// NewMockClient creates a new mock GCP client
func NewMockClient(projectID string) *MockClient {
	return &MockClient{
		projectID: projectID,
		secrets:   make(map[string]string),
		errors:    make(map[string]error),
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

	// Simulate secret already exists error if configured
	if _, exists := m.secrets[name]; exists && m.errors["CreateSecret"] != nil {
		// Just update the value
		m.secrets[name] = value
		return nil
	}

	m.secrets[name] = value
	return nil
}

// GetSecret mocks the GetSecret method
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

// DeleteSecret mocks the DeleteSecret method
func (m *MockClient) DeleteSecret(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["DeleteSecret"]; err != nil {
		return err
	}

	delete(m.secrets, name)
	return nil
}

// Close mocks the Close method
func (m *MockClient) Close() error {
	return nil
}