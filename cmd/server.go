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
	registerAgentConfigFlags(serverCmd.Flags())
}

func runServer(ctx context.Context) error {
	agent, err := initAgent(ctx, agentConfig.Repository, agentConfig.Env, agentConfig.Token, agentConfig.HooksPath)
	if err != nil {
		return err
	}
	return agent.Run(ctx, true, agentConfig.SleepTime)
}
