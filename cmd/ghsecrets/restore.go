package ghsecrets

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tom-023/ghsecrets/internal/aws"
	"github.com/tom-023/ghsecrets/internal/github"
)

var (
	restoreBackup string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore GitHub Secrets from backup",
	Long:  `Restore GitHub Secrets from AWS Secrets Manager or GCP Secret Manager`,
	RunE:  runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Backup source flag (consistent with push command)
	restoreCmd.Flags().StringVarP(&restoreBackup, "backup", "b", "", "Backup source to restore from (aws, gcp)")

	// Repository flags
	restoreCmd.Flags().String("owner", "", "GitHub repository owner")
	restoreCmd.Flags().String("repo", "", "GitHub repository name")

	// AWS specific flags
	restoreCmd.Flags().String("aws-region", "us-east-1", "AWS region")
	restoreCmd.Flags().String("aws-profile", "", "AWS profile name")

	// GCP specific flags
	restoreCmd.Flags().String("gcp-project", "", "GCP project ID")

	// Bind flags to viper
	viper.BindPFlag("github.owner", restoreCmd.Flags().Lookup("owner"))
	viper.BindPFlag("github.repo", restoreCmd.Flags().Lookup("repo"))
	viper.BindPFlag("aws.region", restoreCmd.Flags().Lookup("aws-region"))
	viper.BindPFlag("aws.profile", restoreCmd.Flags().Lookup("aws-profile"))
	viper.BindPFlag("gcp.project", restoreCmd.Flags().Lookup("gcp-project"))
}

func runRestore(cmd *cobra.Command, args []string) error {
	// Validate backup source
	if restoreBackup == "" {
		return fmt.Errorf("backup source must be specified with -b flag (aws or gcp)")
	}

	switch restoreBackup {
	case "aws":
		return runRestoreAWS(cmd, args)
	case "gcp":
		return fmt.Errorf("GCP restore is not yet implemented")
	default:
		return fmt.Errorf("invalid backup source: %s (must be aws or gcp)", restoreBackup)
	}
}

func runRestoreAWS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Get AWS configuration
	awsRegion := viper.GetString("aws.region")
	awsProfile := viper.GetString("aws.profile")
	awsSecretName := viper.GetString("aws.secret_name")

	if awsSecretName == "" {
		return fmt.Errorf("AWS secret name must be configured in ghsecrets.yaml")
	}

	// Get GitHub configuration
	githubOwner := viper.GetString("github.owner")
	githubRepo := viper.GetString("github.repo")
	githubToken := viper.GetString("github.token")

	if githubOwner == "" || githubRepo == "" {
		return fmt.Errorf("GitHub owner and repo must be specified")
	}

	if githubToken == "" {
		githubToken = os.Getenv("GITHUB_TOKEN")
		if githubToken == "" {
			return fmt.Errorf("GitHub token must be set in config or GITHUB_TOKEN environment variable")
		}
	}

	// Create AWS client
	awsOpts := aws.ClientOptions{
		Region:  awsRegion,
		Profile: awsProfile,
	}
	awsClient, err := aws.NewClientWithOptions(awsOpts)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Create GitHub client
	githubClient := github.NewClient(githubToken, githubOwner, githubRepo)

	// Create AWS JSON client
	jsonClient := aws.NewJSONClient(awsClient, awsSecretName)

	// Get all keys from AWS
	keys, err := jsonClient.GetAllKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve secrets from AWS: %w", err)
	}

	if len(keys) == 0 {
		fmt.Println("No secrets found in AWS Secrets Manager")
		return nil
	}

	// Restore each secret to GitHub
	fmt.Printf("Restoring %d secrets from AWS to GitHub repository %s/%s\n", len(keys), githubOwner, githubRepo)
	
	successCount := 0
	for key, value := range keys {
		fmt.Printf("Restoring secret: %s... ", key)
		
		err := githubClient.CreateOrUpdateSecret(ctx, key, value)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		
		fmt.Println("OK")
		successCount++
	}

	fmt.Printf("\nRestore complete: %d/%d secrets successfully restored\n", successCount, len(keys))
	
	if successCount < len(keys) {
		return fmt.Errorf("some secrets failed to restore")
	}

	return nil
}