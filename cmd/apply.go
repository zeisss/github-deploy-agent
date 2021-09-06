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
	registerAgentConfigFlags(applyCmd.Flags())
}

func runApply(ctx context.Context) error {
	agent, err := initAgent(ctx, agentConfig.Repository, agentConfig.Env, agentConfig.Token, agentConfig.HooksPath)
	if err != nil {
		return err
	}
	return agent.Run(ctx, false, agentConfig.SleepTime)
}
