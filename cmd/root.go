package cmd

import (
	"github.com/jdambly/kitter/cmd/client"
	"github.com/jdambly/kitter/cmd/server"
	"github.com/spf13/cobra"
	"os"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(info VersionInfo) {
	rootCmd := newRootCmd(info)
	if err := rootCmd.Execute(); err != nil {
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
