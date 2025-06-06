package ghsecrets

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tom-023/ghsecrets/internal/auth"
	"github.com/tom-023/ghsecrets/internal/aws"
	"github.com/tom-023/ghsecrets/internal/gcp"
	"github.com/tom-023/ghsecrets/internal/github"
	"golang.org/x/term"
)

var (
	key        string
	value      string
	backup     string
	owner      string
	repo       string
	region     string
	awsProfile string
	project    string
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a secret to GitHub and optionally backup to cloud",
	Long: `Push a secret to GitHub Secrets and optionally backup to
AWS Secrets Manager or GCP Secret Manager.

If key or value are not provided via flags, you will be prompted to enter them.
The value input will be hidden for security.

Example:
  ghsecrets push -k API_KEY -v "secret-value" -b aws
  ghsecrets push -k DATABASE_URL -b gcp  # Will prompt for value
  ghsecrets push -k TOKEN  # Will prompt for value
  ghsecrets push  # Will prompt for both key and value`,
	RunE: runPush,
}

func init() {
	rootCmd.AddCommand(pushCmd)

	pushCmd.Flags().StringVarP(&key, "key", "k", "", "Secret key name (will prompt if not provided)")
	pushCmd.Flags().StringVarP(&value, "value", "v", "", "Secret value (will prompt if not provided)")
	pushCmd.Flags().StringVarP(&backup, "backup", "b", "", "Backup destination: aws, gcp, or none")
	pushCmd.Flags().StringVarP(&owner, "owner", "o", "", "GitHub repository owner")
	pushCmd.Flags().StringVarP(&repo, "repo", "r", "", "GitHub repository name")
	pushCmd.Flags().StringVar(&region, "aws-region", "us-east-1", "AWS region for Secrets Manager")
	pushCmd.Flags().StringVar(&awsProfile, "aws-profile", "", "AWS profile to use (from ~/.aws/credentials)")
	pushCmd.Flags().StringVar(&project, "gcp-project", "", "GCP project ID")

	viper.BindPFlag("github.owner", pushCmd.Flags().Lookup("owner"))
	viper.BindPFlag("github.repo", pushCmd.Flags().Lookup("repo"))
	viper.BindPFlag("aws.region", pushCmd.Flags().Lookup("aws-region"))
	viper.BindPFlag("aws.profile", pushCmd.Flags().Lookup("aws-profile"))
	viper.BindPFlag("gcp.project", pushCmd.Flags().Lookup("gcp-project"))
}

func runPush(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get GitHub token
	ghToken, err := auth.GetGitHubToken(viper.GetString("github.token"))
	if err != nil {
		return err
	}

	if owner == "" {
		owner = viper.GetString("github.owner")
	}
	if repo == "" {
		repo = viper.GetString("github.repo")
	}

	if owner == "" || repo == "" {
		return fmt.Errorf("GitHub owner and repo must be specified via flags or config file")
	}

	// If key is not provided, prompt for it
	if key == "" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter secret key name: ")
		keyInput, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read secret key: %w", err)
		}
		key = strings.TrimSpace(keyInput)
		
		// Verify the key is not empty
		if key == "" {
			return fmt.Errorf("secret key cannot be empty")
		}
	}

	// If value is not provided, prompt for it
	if value == "" {
		fmt.Printf("Enter value for secret '%s': ", key)
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read secret value: %w", err)
		}
		fmt.Println() // New line after password input
		value = string(bytePassword)
		
		// Verify the value is not empty
		if value == "" {
			return fmt.Errorf("secret value cannot be empty")
		}
	}

	// Handle backup first if specified
	if backup != "" && backup != "none" {
		fmt.Printf("Creating backup for secret '%s'...\n", key)
		switch strings.ToLower(backup) {
		case "aws":
			if err := backupToAWS(ctx, key, value); err != nil {
				return fmt.Errorf("failed to backup to AWS: %w", err)
			}
			fmt.Println("✓ Successfully backed up to AWS Secrets Manager")

		case "gcp":
			if err := backupToGCP(ctx, key, value); err != nil {
				return fmt.Errorf("failed to backup to GCP: %w", err)
			}
			fmt.Println("✓ Successfully backed up to GCP Secret Manager")

		default:
			return fmt.Errorf("invalid backup destination: %s (use 'aws', 'gcp', or 'none')", backup)
		}
	}

	// Push to GitHub Secrets after successful backup (or if no backup specified)
	fmt.Printf("Pushing secret '%s' to GitHub repository %s/%s...\n", key, owner, repo)
	ghClient := github.NewClient(ghToken, owner, repo)
	if err := ghClient.CreateOrUpdateSecret(ctx, key, value); err != nil {
		return fmt.Errorf("failed to push to GitHub: %w", err)
	}
	fmt.Println("✓ Successfully pushed to GitHub Secrets")

	return nil
}

func backupToAWS(ctx context.Context, key, value string) error {
	awsRegion := viper.GetString("aws.region")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	awsProfile := viper.GetString("aws.profile")
	awsSecretName := viper.GetString("aws.secret_name")
	if awsSecretName == "" {
		// Default secret name if not specified
		awsSecretName = fmt.Sprintf("github-secrets-%s-%s", owner, repo)
	}
	
	awsClient, err := aws.NewClientWithOptions(aws.ClientOptions{
		Region:  awsRegion,
		Profile: awsProfile,
	})
	if err != nil {
		return err
	}

	// Use JSON client to store multiple keys in a single secret
	jsonClient := aws.NewJSONClient(awsClient, awsSecretName)
	return jsonClient.AddOrUpdateKey(ctx, key, value)
}

func backupToGCP(ctx context.Context, key, value string) error {
	gcpProject := viper.GetString("gcp.project")
	if gcpProject == "" {
		return fmt.Errorf("GCP project ID not specified. Use --gcp-project flag or configure in .ghsecrets.yaml")
	}

	gcpCreds := viper.GetString("gcp.credentials_path")
	
	gcpClient, err := gcp.NewClient(gcpProject, gcpCreds)
	if err != nil {
		return err
	}
	defer gcpClient.Close()

	return gcpClient.CreateOrUpdateSecret(ctx, key, value)
}