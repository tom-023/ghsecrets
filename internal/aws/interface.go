package aws

import "context"

// SecretClient defines the interface for secret operations
type SecretClient interface {
	CreateOrUpdateSecret(ctx context.Context, name, value, description string) error
	GetSecret(ctx context.Context, name string) (string, error)
}