package aws

import (
	"context"
	"encoding/json"
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