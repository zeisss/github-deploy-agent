package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = cobra.Command{
	Use:   "github-deploy-agent",
	Short: "Agent to run custom scripts triggered via GitHub Deployment API",
}

func Execute() error {
	return rootCmd.Execute()
}

var agentConfig struct {
	Repository string
	Env        string
	Token      string
	SleepTime  time.Duration
	HooksPath  string
}

func init() {
	rootCmd.PersistentFlags().StringVar(&agentConfig.HooksPath, "tasks.path", "./hooks", "Path to directory with tasks to be executed")
	rootCmd.PersistentFlags().StringVar(&agentConfig.Repository, "repository", "", "The repository to deploy for - owner/repo")
	rootCmd.PersistentFlags().StringVar(&agentConfig.Env, "env", "production", "The environment to act on")
	rootCmd.PersistentFlags().StringVar(&agentConfig.Token, "token", "", "API token for github")
	rootCmd.PersistentFlags().DurationVar(&agentConfig.SleepTime, "sleep", 60*time.Second, "Sleep time between checks to github api")
}
