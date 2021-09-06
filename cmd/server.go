package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var serverCmd = cobra.Command{
	Use:   "server",
	Short: "Periodically check for waiting deployments and apply them.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(&serverCmd)
}

func runServer(ctx context.Context) error {
	agent := initAgent(ctx, agentConfig.Repository, agentConfig.Env, agentConfig.Token)
	return agent.Run(ctx, true, agentConfig.SleepTime)
}
