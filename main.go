package main

// Import necessary packages
import (
	"os"

	"github.com/netSkopePlatformEng/k8s-jitter/cmd"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// These variables are populated by goreleaser when the binary is built.
// version represents the current version of the application.
// commit represents the git commit that the application was built from.
// date represents the date when the application was built.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Set the time field format for the logger to Unix time, which is the number of seconds elapsed since January 1, 1970 UTC.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure the global logger to include the caller (file and line number) in log messages.
	log.Logger = log.With().Caller().Logger()

	// Configure the global logger to write log messages to stderr in a human-friendly format.
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Execute the main command of the application, passing in the version info.
	cmd.Execute(cmd.VersionInfo{Version: version, Commit: commit, Date: date})
}
