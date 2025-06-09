package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONClient_AddOrUpdateKey(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "test-secret")

	// First create the secret (empty JSON)
	err := mockClient.CreateOrUpdateSecret(ctx, "test-secret", "{}", "test secret")
	require.NoError(t, err)

	// Test adding first key
	err = jsonClient.AddOrUpdateKey(ctx, "key1", "value1")
	require.NoError(t, err)

	// Verify the secret contains the key
	secretJSON, err := mockClient.GetSecret(ctx, "test-secret")
	require.NoError(t, err)

	var data map[string]string
	err = json.Unmarshal([]byte(secretJSON), &data)
	require.NoError(t, err)
	assert.Equal(t, "value1", data["key1"])

	// Test adding second key
	err = jsonClient.AddOrUpdateKey(ctx, "key2", "value2")
	require.NoError(t, err)

	// Verify both keys exist
	secretJSON, err = mockClient.GetSecret(ctx, "test-secret")
	require.NoError(t, err)

	err = json.Unmarshal([]byte(secretJSON), &data)
	require.NoError(t, err)
	assert.Equal(t, "value1", data["key1"])
	assert.Equal(t, "value2", data["key2"])

	// Test updating existing key
	err = jsonClient.AddOrUpdateKey(ctx, "key1", "updated-value1")
	require.NoError(t, err)

	// Verify key was updated
	secretJSON, err = mockClient.GetSecret(ctx, "test-secret")
	require.NoError(t, err)

	err = json.Unmarshal([]byte(secretJSON), &data)
	require.NoError(t, err)
	assert.Equal(t, "updated-value1", data["key1"])
	assert.Equal(t, "value2", data["key2"])
}

func TestJSONClient_GetKey(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "test-secret")

	// Setup test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	jsonData, _ := json.Marshal(testData)
	mockClient.CreateOrUpdateSecret(ctx, "test-secret", string(jsonData), "test")

	// Test getting existing key
	value, err := jsonClient.GetKey(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", value)

	// Test getting non-existent key
	_, err = jsonClient.GetKey(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key non-existent not found")
}

func TestJSONClient_GetAllKeys(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "test-secret")

	// Setup test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	jsonData, _ := json.Marshal(testData)
	mockClient.CreateOrUpdateSecret(ctx, "test-secret", string(jsonData), "test")

	// Test getting all keys
	allKeys, err := jsonClient.GetAllKeys(ctx)
	require.NoError(t, err)
	assert.Equal(t, testData, allKeys)
}

func TestJSONClient_NonExistentSecret(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "non-existent-secret")

	// Test adding key to non-existent secret (should return error)
	err := jsonClient.AddOrUpdateKey(ctx, "key1", "value1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AWS Secrets Manager secret 'non-existent-secret' not found")
}

func TestJSONClient_InvalidJSONFormat(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "invalid-json-secret")

	// Create a secret with invalid JSON
	mockClient.CreateOrUpdateSecret(ctx, "invalid-json-secret", "not-a-json-string", "test")

	// Test adding key to secret with invalid JSON
	err := jsonClient.AddOrUpdateKey(ctx, "key1", "value1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret 'invalid-json-secret' exists but is not in valid JSON format")
}

func TestJSONClient_AuthenticationErrors(t *testing.T) {
	ctx := context.Background()
	mockClient := NewMockClient()
	jsonClient := NewJSONClient(mockClient, "test-secret")

	tests := []struct {
		name          string
		errorMessage  string
		expectedError string
	}{
		{
			name:          "Expired token error",
			errorMessage:  "ExpiredTokenException: The security token included in the request is expired",
			expectedError: "AWS authentication error",
		},
		{
			name:          "Invalid token error",
			errorMessage:  "InvalidTokenException: The security token included in the request is invalid",
			expectedError: "AWS authentication error",
		},
		{
			name:          "No credential providers",
			errorMessage:  "NoCredentialProviders: no valid providers in chain",
			expectedError: "AWS authentication error",
		},
		{
			name:          "Access denied",
			errorMessage:  "AccessDeniedException: User is not authorized to perform this action",
			expectedError: "AWS authentication error",
		},
		{
			name:          "SSO token expired",
			errorMessage:  "token has expired, refresh with aws sso login",
			expectedError: "AWS authentication error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set error for GetSecret
			mockClient.SetError("GetSecret", fmt.Errorf(tt.errorMessage))

			// Test AddOrUpdateKey
			err := jsonClient.AddOrUpdateKey(ctx, "key1", "value1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)
			assert.Contains(t, err.Error(), "aws sso login")

			// Test GetKey
			_, err = jsonClient.GetKey(ctx, "key1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Test GetAllKeys
			_, err = jsonClient.GetAllKeys(ctx)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedError)

			// Clear error for next test
			mockClient.SetError("GetSecret", nil)
		})
	}
}