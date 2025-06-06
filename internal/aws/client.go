package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Client struct {
	client *secretsmanager.Client
	region string
}

func NewClient(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &Client{
		client: client,
		region: region,
	}, nil
}

func (c *Client) CreateOrUpdateSecret(ctx context.Context, name, value, description string) error {
	// First, try to create the secret
	_, err := c.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(name),
		SecretString: aws.String(value),
		Description:  aws.String(description),
	})

	if err != nil {
		// If secret already exists, update it
		if isSecretExistsError(err) {
			_, err = c.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
				SecretId:     aws.String(name),
				SecretString: aws.String(value),
				Description:  aws.String(description),
			})
			if err != nil {
				return fmt.Errorf("failed to update secret: %w", err)
			}
		} else {
			return fmt.Errorf("failed to create secret: %w", err)
		}
	}

	return nil
}

func (c *Client) GetSecret(ctx context.Context, name string) (string, error) {
	result, err := c.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	if result.SecretString != nil {
		return *result.SecretString, nil
	}

	return "", fmt.Errorf("secret value is empty")
}

func (c *Client) DeleteSecret(ctx context.Context, name string) error {
	_, err := c.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(name),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}

func isSecretExistsError(err error) bool {
	// Check if error indicates that secret already exists
	if err != nil {
		return err.Error() == "ResourceExistsException: The operation failed because the secret version already exists." ||
			err.Error() == "ResourceExistsException: A resource with the ID you requested already exists."
	}
	return false
}