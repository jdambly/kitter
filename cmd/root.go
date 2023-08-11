package cmd

import (
	"context"
	"os"

	"github.com/jdambly/kitter/cmd/client"
	"github.com/jdambly/kitter/cmd/server"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context, info VersionInfo) {
	rootCmd := newRootCmd(info)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

// newRootCmd adds all the sub commands
func newRootCmd(info VersionInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kitter",
		Short: "kitter k8s jitter",
	}
	cmd.AddCommand(
		newVersionCmd(info),
		client.NewCmd(),
		server.NewCmd(),
	)
	return cmd
}
