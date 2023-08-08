package client

import (
	"github.com/spf13/cobra"
)

// NewCmd
func NewCmd() *cobra.Command {
	// create the "client" command
	cmd := &cobra.Command{
		Use:   "client",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the value of the "host" flag
			host, _ := cmd.Flags().GetString("host")
			return worker(host)
		},
	}
	// Add the "host" flag to the "client" command.
	cmd.Flags().StringP("host", "h", "", "Host to connect to")

	// Return the new command.
	return cmd
}

func worker(host string) error {
	return nil
}
