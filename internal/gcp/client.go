package gcp

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/option"
)

type Client struct {
	client    *secretmanager.Client
	projectID string
}

func NewClient(projectID string, credentialsPath string) (*Client, error) {
	ctx := context.Background()

	var opts []option.ClientOption
	if credentialsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsPath))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &Client{
		client:    client,
		projectID: projectID,
	}, nil
}

func (c *Client) CreateOrUpdateSecret(ctx context.Context, name, value string) error {
	parent := fmt.Sprintf("projects/%s", c.projectID)
	secretID := name

	// First, try to create the secret
	createReq := &secretmanagerpb.CreateSecretRequest{
		Parent:   parent,
		SecretId: secretID,
		Secret: &secretmanagerpb.Secret{
			Replication: &secretmanagerpb.Replication{
				Replication: &secretmanagerpb.Replication_Automatic_{
					Automatic: &secretmanagerpb.Replication_Automatic{},
				},
			},
		},
	}

	secret, err := c.client.CreateSecret(ctx, createReq)
	if err != nil {
		// If secret already exists, get it
		if isSecretExistsError(err) {
			getReq := &secretmanagerpb.GetSecretRequest{
				Name: fmt.Sprintf("projects/%s/secrets/%s", c.projectID, secretID),
			}
			secret, err = c.client.GetSecret(ctx, getReq)
			if err != nil {
				return fmt.Errorf("failed to get existing secret: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	// Add a new version with the value
	addVersionReq := &secretmanagerpb.AddSecretVersionRequest{
		Parent: secret.Name,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(value),
		},
	}

	_, err = c.client.AddSecretVersion(ctx, addVersionReq)
	if err != nil {
		return fmt.Errorf("failed to add secret version: %w", err)
	}

	return nil
}

func (c *Client) GetSecret(ctx context.Context, name string) (string, error) {
	accessReq := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fmt.Sprintf("projects/%s/secrets/%s/versions/latest", c.projectID, name),
	}

	result, err := c.client.AccessSecretVersion(ctx, accessReq)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func isSecretExistsError(err error) bool {
	// Check if error indicates that secret already exists
	if err != nil {
		return err.Error() == "rpc error: code = AlreadyExists desc = Secret [projects/*/secrets/*] already exists."
	}
	return false
}