package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRootCmdElastic(t *testing.T) {
	info := VersionInfo{Version: "version", Commit: "commit", Date: "date"}
	rootCmd := newRootCmd(info)
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b) // capture the outPut into a string buffer
	rootCmd.SetArgs([]string{"elastic"})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}
