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

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore GitHub Secrets from backup",
	Long:  `Restore GitHub Secrets from AWS Secrets Manager or GCP Secret Manager`,
}

var restoreAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "Restore GitHub Secrets from AWS Secrets Manager",
	Long:  `Restore all GitHub Secrets from AWS Secrets Manager JSON backup`,
	RunE:  runRestoreAWS,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.AddCommand(restoreAWSCmd)

	// AWS specific flags
	restoreAWSCmd.Flags().String("aws-region", "us-east-1", "AWS region")
	restoreAWSCmd.Flags().String("aws-profile", "", "AWS profile name")
	restoreAWSCmd.Flags().String("owner", "", "GitHub repository owner")
	restoreAWSCmd.Flags().String("repo", "", "GitHub repository name")

	// Bind flags to viper
	viper.BindPFlag("aws.region", restoreAWSCmd.Flags().Lookup("aws-region"))
	viper.BindPFlag("aws.profile", restoreAWSCmd.Flags().Lookup("aws-profile"))
	viper.BindPFlag("github.owner", restoreAWSCmd.Flags().Lookup("owner"))
	viper.BindPFlag("github.repo", restoreAWSCmd.Flags().Lookup("repo"))
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