package ghsecrets

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "ghsecrets",
		Short: "A CLI tool to manage GitHub Secrets with cloud backup",
		Long: `ghsecrets is a CLI tool that allows you to manage GitHub Secrets
while automatically backing them up to cloud secret management services like
AWS Secrets Manager and GCP Secret Manager.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ghsecrets.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Look for config file in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("ghsecrets")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("GHSECRETS")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}