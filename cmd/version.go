package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"runtime"
)

// VersionInfo is information about the version of nsk.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}

func newVersionCmd(info VersionInfo) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "Display kitter version information.",
		Long:         "Display kitter version information.",
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			printVersionInfo(info)
		},
	}
}

func printVersionInfo(info VersionInfo) {
	fmt.Printf("kitter version=%q commit=%q go=%q build_date=%q\n", info.Version, info.Commit, runtime.Version(), info.Date)
}
