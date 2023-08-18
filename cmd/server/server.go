package server

import (
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewCmd
func NewCmd() *cobra.Command {
	var port string
	// create the "server" command
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// make a channel to check if the tcp server is ready or not
			readyCh := make(chan struct{})
			// create the new server
			srv, err := netapi.NewServer("tcp", ":"+port)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to create server")
			}

			// Start the server
			go func() {
				log.Info().Msg("Starting TCP server on tcp addr -> :" + port)
				if err := srv.Run(readyCh); err != nil {
					log.Fatal().Err(err).Msg("failed to start server")
				}
			}()
			<-readyCh
			// block with select and wait for channel returns
			<-cmd.Context().Done()
			log.Info().Msg("shutting down")
			return nil
		},
	}

	// add the "port" flag to the "server" command
	cmd.Flags().StringVarP(&port, "port", "p", "5102", "TCP port to listen on")

	// Return the new command
	return cmd
}
