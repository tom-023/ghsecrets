package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// JSONClient wraps the AWS client to store multiple key-value pairs in a single secret
type JSONClient struct {
	client     SecretClient
	secretName string
}

// NewJSONClient creates a new client that stores secrets as JSON
func NewJSONClient(client SecretClient, secretName string) *JSONClient {
	return &JSONClient{
		client:     client,
		secretName: secretName,
	}
}

// AddOrUpdateKey adds or updates a key-value pair in the JSON secret
func (j *JSONClient) AddOrUpdateKey(ctx context.Context, key, value string) error {
	// First check if the secret exists
	existingJSON, err := j.client.GetSecret(ctx, j.secretName)
	if err != nil {
		return j.wrapGetSecretError(err)
	}
	
	// Parse existing secret data
	secretData := make(map[string]string)
	if existingJSON != "" {
		if err := json.Unmarshal([]byte(existingJSON), &secretData); err != nil {
			// If unmarshaling fails, it might not be JSON format
			// Return error instead of starting fresh
			return fmt.Errorf("secret '%s' exists but is not in valid JSON format: %w", j.secretName, err)
		}
	}
	
	// Add or update the key
	secretData[key] = value
	
	// Marshal back to JSON
	updatedJSON, err := json.MarshalIndent(secretData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secret data: %w", err)
	}
	
	// Update the secret
	description := fmt.Sprintf("GitHub Secrets backup (JSON format)")
	return j.client.CreateOrUpdateSecret(ctx, j.secretName, string(updatedJSON), description)
}

// GetKey retrieves a specific key from the JSON secret
func (j *JSONClient) GetKey(ctx context.Context, key string) (string, error) {
	existingJSON, err := j.client.GetSecret(ctx, j.secretName)
	if err != nil {
		return "", j.wrapGetSecretError(err)
	}
	
	var secretData map[string]string
	if err := json.Unmarshal([]byte(existingJSON), &secretData); err != nil {
		return "", fmt.Errorf("failed to unmarshal secret data: %w", err)
	}
	
	value, exists := secretData[key]
	if !exists {
		return "", fmt.Errorf("key %s not found in secret", key)
	}
	
	return value, nil
}

// GetAllKeys retrieves all key-value pairs from the JSON secret
func (j *JSONClient) GetAllKeys(ctx context.Context) (map[string]string, error) {
	existingJSON, err := j.client.GetSecret(ctx, j.secretName)
	if err != nil {
		return nil, j.wrapGetSecretError(err)
	}
	
	var secretData map[string]string
	if err := json.Unmarshal([]byte(existingJSON), &secretData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret data: %w", err)
	}
	
	return secretData, nil
}

// wrapGetSecretError wraps GetSecret errors with more meaningful messages
func (j *JSONClient) wrapGetSecretError(err error) error {
	// First check for common authentication/authorization errors in the error message
	// This must come before ResourceNotFoundException check because AWS may return
	// ResourceNotFoundException even when the real issue is authentication
	errStr := err.Error()
	if strings.Contains(errStr, "ExpiredToken") || 
	   strings.Contains(errStr, "InvalidToken") ||
	   strings.Contains(errStr, "NoCredentialProviders") ||
	   strings.Contains(errStr, "UnauthorizedException") ||
	   strings.Contains(errStr, "AccessDenied") ||
	   strings.Contains(errStr, "no valid credential") ||
	   strings.Contains(errStr, "failed to retrieve credentials") ||
	   strings.Contains(errStr, "token has expired") {
		return fmt.Errorf("AWS authentication error: %w. Please check your AWS credentials or run 'aws sso login' if using SSO", err)
	}
	
	// Check if it's a resource not found error (only after ruling out auth issues)
	var resourceNotFoundErr *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundErr) {
		return fmt.Errorf("AWS Secrets Manager secret '%s' not found. Please create it first in AWS console or specify a different secret_name in config", j.secretName)
	}
	
	// For any other error, return it as-is
	return fmt.Errorf("failed to access AWS Secrets Manager: %w", err)
}