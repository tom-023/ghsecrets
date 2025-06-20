package aws

import (
	"context"
	"fmt"
	"sync"
	
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// MockClient is a mock implementation of AWS Secrets Manager client for testing
type MockClient struct {
	mu      sync.Mutex
	secrets map[string]string
	errors  map[string]error
}

// NewMockClient creates a new mock AWS client
func NewMockClient() *MockClient {
	return &MockClient{
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
func (m *MockClient) CreateOrUpdateSecret(ctx context.Context, name, value, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.errors["CreateOrUpdateSecret"]; err != nil {
		return err
	}

	// Simulate "already exists" error if secret exists and we're in create mode
	if _, exists := m.secrets[name]; exists && m.errors["CreateSecret"] != nil {
		// Update the secret instead
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
		// Return AWS SDK ResourceNotFoundException for consistency
		return "", &types.ResourceNotFoundException{
			Message: &[]string{fmt.Sprintf("Secrets Manager can't find the specified secret.")}[0],
		}
	}

	return value, nil
}

