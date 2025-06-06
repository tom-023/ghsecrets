package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

type Client struct {
	client *secretsmanager.Client
	region string
}

// ClientOptions contains options for creating an AWS client
type ClientOptions struct {
	Region  string
	Profile string
}

// NewClient creates a new AWS Secrets Manager client with the specified region
func NewClient(region string) (*Client, error) {
	return NewClientWithOptions(ClientOptions{Region: region})
}

// NewClientWithOptions creates a new AWS Secrets Manager client with options
func NewClientWithOptions(opts ClientOptions) (*Client, error) {
	configOpts := []func(*config.LoadOptions) error{
		config.WithRegion(opts.Region),
	}

	// Add profile if specified
	if opts.Profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(opts.Profile))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), configOpts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &Client{
		client: client,
		region: opts.Region,
	}, nil
}

func (c *Client) CreateOrUpdateSecret(ctx context.Context, name, value, description string) error {
	// Try to update existing secret first
	_, err := c.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(name),
		SecretString: aws.String(value),
		Description:  aws.String(description),
	})

	if err != nil {
		// Check if secret doesn't exist
		var resourceNotFoundErr *types.ResourceNotFoundException
		if errors.As(err, &resourceNotFoundErr) {
			// Secret doesn't exist, try to create it
			_, createErr := c.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
				Name:         aws.String(name),
				SecretString: aws.String(value),
				Description:  aws.String(description),
			})
			if createErr != nil {
				return fmt.Errorf("failed to create secret: %w", createErr)
			}
			return nil
		}
		return fmt.Errorf("failed to update secret: %w", err)
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

func isSecretExistsError(err error) bool {
	// Check if error indicates that secret already exists
	var resourceExistsErr *types.ResourceExistsException
	return errors.As(err, &resourceExistsErr)
}