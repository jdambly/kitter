package server

import "github.com/spf13/cobra"

// NewCmd
func NewCmd() *cobra.Command {
	// create the "server" command
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the value of the "port" flag
			port, _ := cmd.Flags().GetInt("port")
			return worker(port)
		},
	}

	// add the "port" flag to the "server" command
	cmd.Flags().IntP("port", "p", 5102, "TCP port to listen on")

	// Return the new command
	return cmd
}

func worker(port int) error {
	return nil
}
