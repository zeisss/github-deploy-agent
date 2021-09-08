package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = cobra.Command{
	Use:   "github-deploy-agent",
	Short: "Agent to run custom scripts triggered via GitHub Deployment API",
}

var createCmd = cobra.Command{
	Use:   "create",
	Short: "Create a resource",
}

func init() {
	rootCmd.AddCommand(&createCmd)
	rootCmd.PersistentFlags().StringVar(&commonConfig.Token, "token", "", "API token for github")
}

func Execute() error {
	return rootCmd.Execute()
}

var commonConfig struct {
	Token string
}
var agentConfig struct {
	Repository string
	Env        string
	SleepTime  time.Duration
	HooksPath  string
}

func registerAgentConfigFlags(flags *pflag.FlagSet) {
	flags.StringVar(&agentConfig.HooksPath, "tasks.path", "./hooks", "Path to directory with tasks to be executed")
	flags.StringVar(&agentConfig.Repository, "repository", "", "The repository to deploy for - owner/repo")
	flags.StringVar(&agentConfig.Env, "env", "testing", "The environment to act on")
	flags.DurationVar(&agentConfig.SleepTime, "sleep", 60*time.Second, "Sleep time between checks to github api")
}
