package ghsecrets

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tom-023/ghsecrets/internal/aws"
	"github.com/tom-023/ghsecrets/internal/gcp"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets from cloud providers",
	Long: `List secrets stored in AWS Secrets Manager or GCP Secret Manager.
Note: GitHub API does not support listing secret values, only secret names.`,
}

var listAWSCmd = &cobra.Command{
	Use:   "aws",
	Short: "List secrets from AWS Secrets Manager",
	RunE:  runListAWS,
}

var listGCPCmd = &cobra.Command{
	Use:   "gcp",
	Short: "List secrets from GCP Secret Manager",
	RunE:  runListGCP,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listAWSCmd)
	listCmd.AddCommand(listGCPCmd)
}

func runListAWS(cmd *cobra.Command, args []string) error {
	awsRegion := viper.GetString("aws.region")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	_, err := aws.NewClient(awsRegion)
	if err != nil {
		return fmt.Errorf("failed to create AWS client: %w", err)
	}

	// Note: AWS SDK doesn't have a simple list secrets method in the current implementation
	// This would need to be implemented in the AWS client
	fmt.Println("AWS Secrets Manager listing functionality to be implemented")

	return nil
}

func runListGCP(cmd *cobra.Command, args []string) error {
	gcpProject := viper.GetString("gcp.project")
	if gcpProject == "" {
		return fmt.Errorf("GCP project ID not specified. Use --gcp-project flag or configure in .ghsecrets.yaml")
	}

	gcpCreds := viper.GetString("gcp.credentials_path")

	gcpClient, err := gcp.NewClient(gcpProject, gcpCreds)
	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}
	defer gcpClient.Close()

	// Note: GCP SDK listing functionality would need to be implemented in the GCP client
	fmt.Println("GCP Secret Manager listing functionality to be implemented")

	return nil
}
