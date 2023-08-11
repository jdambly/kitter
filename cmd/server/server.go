package server

import (
	"fmt"
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

// NewCmd
func NewCmd() *cobra.Command {
	// create the "server" command
	cmd := &cobra.Command{
		Use:   "server",
		Short: "start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error = nil
			// get the value of the "port" flag
			port, _ := cmd.Flags().GetString("port")

			// create the channels
			errServer := make(chan error)
			stopChan := make(chan os.Signal, 1)
			signal.Notify(stopChan, os.Interrupt)
			signal.Notify(stopChan, syscall.SIGTERM)

			// create the new server
			srv, err := netapi.NewServer("tcp", ":"+port)
			if err != nil {
				log.Error().Err(err).Msg("failed to create server")
			}

			// Start the server
			go func() {
				log.Info().Msg("Starting TCP server on tcp addr -> :" + port)
				errServer <- srv.Run()
			}()

			// block with select and wait for channel returns
			select {
			case errS := <-errServer:
				if err != nil {
					log.Error().Err(errS).Msg("TCP server error")
				}
			case s := <-stopChan:
				log.Info().Msg(fmt.Sprintf("signal %s received", s.String()))
			}

			return err
		},
	}

	// add the "port" flag to the "server" command
	cmd.Flags().StringP("port", "p", "5102", "TCP port to listen on")

	// Return the new command
	return cmd
}
