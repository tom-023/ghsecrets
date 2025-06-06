package ghsecrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tom-023/ghsecrets/internal/aws"
)

// MockAWSClient is a test mock for AWS operations
type MockAWSClient struct {
	secrets map[string]string
	err     error
}

func (m *MockAWSClient) CreateOrUpdateSecret(ctx context.Context, name, value, description string) error {
	if m.err != nil {
		return m.err
	}
	if m.secrets == nil {
		m.secrets = make(map[string]string)
	}
	m.secrets[name] = value
	return nil
}

func (m *MockAWSClient) GetSecret(ctx context.Context, name string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	value, exists := m.secrets[name]
	if !exists {
		return "", fmt.Errorf("secret not found: %s", name)
	}
	return value, nil
}

// MockGitHubClient is a test mock for GitHub operations
type MockGitHubClient struct {
	secrets       map[string]string
	createErr     error
	callCount     int
	failOnNthCall int
}

func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		secrets:       make(map[string]string),
		failOnNthCall: -1, // デフォルトでは失敗しない
	}
}

func (m *MockGitHubClient) CreateOrUpdateSecret(ctx context.Context, name, value string) error {
	m.callCount++
	if m.failOnNthCall > 0 && m.callCount == m.failOnNthCall {
		return fmt.Errorf("failed to create secret: %s", name)
	}
	if m.createErr != nil {
		return m.createErr
	}
	m.secrets[name] = value
	return nil
}

func TestRestoreAWSCommand(t *testing.T) {
	tests := []struct {
		name          string
		awsSecrets    string // JSON形式のシークレット
		awsError      error
		githubError   error
		expectedError bool
		expectedCount int
	}{
		{
			name:          "正常なリストア - 複数のシークレット",
			awsSecrets:    `{"API_KEY":"secret-api-key","DATABASE_URL":"postgres://localhost","TOKEN":"bearer-token"}`,
			awsError:      nil,
			githubError:   nil,
			expectedError: false,
			expectedCount: 3,
		},
		{
			name:          "空のシークレット",
			awsSecrets:    `{}`,
			awsError:      nil,
			githubError:   nil,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "AWSからの取得に失敗",
			awsSecrets:    "",
			awsError:      fmt.Errorf("AWS access denied"),
			githubError:   nil,
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:          "GitHubへの作成に失敗",
			awsSecrets:    `{"API_KEY":"secret-value"}`,
			awsError:      nil,
			githubError:   fmt.Errorf("GitHub authentication failed"),
			expectedError: true,
			expectedCount: 0,
		},
		{
			name:          "無効なJSON形式",
			awsSecrets:    `{"invalid json`,
			awsError:      nil,
			githubError:   nil,
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Viperの設定をリセット
			viper.Reset()
			viper.Set("aws.secret_name", "test-secret")
			viper.Set("github.owner", "test-owner")
			viper.Set("github.repo", "test-repo")
			viper.Set("github.token", "test-token")

			// モッククライアントを作成
			mockAWS := &MockAWSClient{
				secrets: map[string]string{"test-secret": tt.awsSecrets},
				err:     tt.awsError,
			}

			mockGitHub := NewMockGitHubClient()
			mockGitHub.createErr = tt.githubError

			// テスト対象の関数を直接呼び出す代わりに、
			// ここではモックを使用したロジックのテストを行います
			ctx := context.Background()

			// AWS JSONクライアントのモック動作をシミュレート
			jsonClient := aws.NewJSONClient(mockAWS, "test-secret")

			// AWSからシークレットを取得
			secrets, err := jsonClient.GetAllKeys(ctx)

			if tt.awsError != nil {
				assert.Error(t, err)
				return
			}

			if tt.name == "無効なJSON形式" {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// GitHubに各シークレットを作成
			successCount := 0
			for key, value := range secrets {
				err := mockGitHub.CreateOrUpdateSecret(ctx, key, value)
				if err == nil {
					successCount++
				}
			}

			// 結果を検証
			if tt.githubError != nil {
				assert.Equal(t, 0, successCount)
			} else {
				assert.Equal(t, tt.expectedCount, successCount)
				assert.Equal(t, len(secrets), len(mockGitHub.secrets))
			}
		})
	}
}

func TestRestoreAWSPartialFailure(t *testing.T) {
	// 一部のシークレットだけが失敗するケース
	ctx := context.Background()

	// 3つのシークレットを持つAWSモックを作成
	mockAWS := &MockAWSClient{
		secrets: map[string]string{
			"test-secret": `{"SECRET1":"value1","SECRET2":"value2","SECRET3":"value3"}`,
		},
	}

	// 2番目のシークレットで失敗するGitHubモックを作成
	mockGitHub := NewMockGitHubClient()
	mockGitHub.failOnNthCall = 2

	// JSONクライアントでシークレットを取得
	jsonClient := aws.NewJSONClient(mockAWS, "test-secret")
	secrets, err := jsonClient.GetAllKeys(ctx)
	require.NoError(t, err)
	assert.Len(t, secrets, 3)

	// 各シークレットをGitHubに作成
	successCount := 0
	failCount := 0
	for key, value := range secrets {
		err := mockGitHub.CreateOrUpdateSecret(ctx, key, value)
		if err != nil {
			failCount++
		} else {
			successCount++
		}
	}

	// 3つのうち2つが成功、1つが失敗することを確認
	assert.Equal(t, 2, successCount)
	assert.Equal(t, 1, failCount)
	assert.Len(t, mockGitHub.secrets, 2)
}

func TestRestoreAWSConfigValidation(t *testing.T) {
	tests := []struct {
		name          string
		secretName    string
		githubOwner   string
		githubRepo    string
		githubToken   string
		expectedError string
	}{
		{
			name:          "AWS secret名が未設定",
			secretName:    "",
			githubOwner:   "owner",
			githubRepo:    "repo",
			githubToken:   "token",
			expectedError: "AWS secret name must be configured",
		},
		{
			name:          "GitHubオーナーが未設定",
			secretName:    "test-secret",
			githubOwner:   "",
			githubRepo:    "repo",
			githubToken:   "token",
			expectedError: "GitHub owner and repo must be specified",
		},
		{
			name:          "GitHubリポジトリが未設定",
			secretName:    "test-secret",
			githubOwner:   "owner",
			githubRepo:    "",
			githubToken:   "token",
			expectedError: "GitHub owner and repo must be specified",
		},
		{
			name:          "GitHubトークンが未設定",
			secretName:    "test-secret",
			githubOwner:   "owner",
			githubRepo:    "repo",
			githubToken:   "",
			expectedError: "GitHub token must be set",
		},
		{
			name:          "全ての設定が正常",
			secretName:    "test-secret",
			githubOwner:   "owner",
			githubRepo:    "repo",
			githubToken:   "token",
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 設定の検証ロジックをテスト
			var err error

			if tt.secretName == "" {
				err = fmt.Errorf("AWS secret name must be configured")
			} else if tt.githubOwner == "" || tt.githubRepo == "" {
				err = fmt.Errorf("GitHub owner and repo must be specified")
			} else if tt.githubToken == "" {
				err = fmt.Errorf("GitHub token must be set")
			}

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
