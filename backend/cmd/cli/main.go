package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "1.0.0"
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mobile-backend-cli",
	Short: "A comprehensive CLI tool for mobile backend management",
	Long: `Mobile Backend CLI is a powerful command-line tool that provides:

- Code generation and scaffolding
- API testing and exploration
- Database management and migrations
- Deployment and environment management
- Testing and debugging utilities
- Documentation generation

Built for developers who want to quickly scaffold, test, and deploy mobile backends.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mobile-backend-cli.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringP("base-url", "u", "http://localhost:8080", "base URL for API calls")
	rootCmd.PersistentFlags().StringP("api-key", "k", "", "API key for authentication")

	// Bind flags to viper
	viper.BindPFlag("base_url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add subcommands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(explorerCmd)
	rootCmd.AddCommand(sdkCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mobile-backend-cli" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mobile-backend-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil && verbose {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	Execute()
}
