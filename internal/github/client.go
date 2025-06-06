package github

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v47/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	owner  string
	repo   string
	token  string
}

func NewClient(token, owner, repo string) *Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &Client{
		client: client,
		owner:  owner,
		repo:   repo,
		token:  token,
	}
}

func (c *Client) CreateOrUpdateSecret(ctx context.Context, name, value string) error {
	publicKey, err := c.getPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}

	encryptedValue, err := encryptSecret(publicKey.GetKey(), value)
	if err != nil {
		return fmt.Errorf("failed to encrypt secret: %w", err)
	}

	_, err = c.client.Actions.CreateOrUpdateRepoSecret(ctx, c.owner, c.repo, &github.EncryptedSecret{
		Name:           name,
		KeyID:          publicKey.GetKeyID(),
		EncryptedValue: encryptedValue,
	})
	if err != nil {
		return fmt.Errorf("failed to create/update secret: %w", err)
	}

	return nil
}

func (c *Client) getPublicKey(ctx context.Context) (*github.PublicKey, error) {
	publicKey, _, err := c.client.Actions.GetRepoPublicKey(ctx, c.owner, c.repo)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func encryptSecret(publicKey, secret string) (string, error) {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	var publicKeyArray [32]byte
	copy(publicKeyArray[:], publicKeyBytes)

	encrypted, err := box.SealAnonymous(nil, []byte(secret), &publicKeyArray, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

type ActionsPublicKey struct {
	KeyID string `json:"key_id"`
	Key   string `json:"key"`
}

func (c *Client) GetPublicKeyManual(ctx context.Context) (*ActionsPublicKey, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/secrets/public-key", c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get public key: %s", body)
	}

	var publicKey ActionsPublicKey
	if err := json.NewDecoder(resp.Body).Decode(&publicKey); err != nil {
		return nil, err
	}

	return &publicKey, nil
}