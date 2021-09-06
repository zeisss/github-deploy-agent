package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var applyCmd = cobra.Command{
	Use:   "apply",
	Short: "Check for waiting deployments, apply then exit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApply(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(&applyCmd)
}

func runApply(ctx context.Context) error {
	agent := initAgent(ctx, agentConfig.Repository, agentConfig.Env, agentConfig.Token)
	return agent.Run(ctx, false, agentConfig.SleepTime)
}
