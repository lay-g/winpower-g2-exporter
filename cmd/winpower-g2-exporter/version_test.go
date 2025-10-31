package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionCmd(t *testing.T) {
	cmd := NewVersionCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Use)
}

func TestGetVersionInfo(t *testing.T) {
	info := getVersionInfo()

	assert.NotNil(t, info)
	assert.NotEmpty(t, info.GoVersion)
	assert.NotEmpty(t, info.Platform)
	assert.NotEmpty(t, info.Compiler)
}

func TestVersionCmdTextOutput(t *testing.T) {
	cmd := NewVersionCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestVersionCmdJSONOutput(t *testing.T) {
	cmd := NewVersionCmd()
	cmd.SetArgs([]string{"--format", "json"})

	err := cmd.Execute()
	assert.NoError(t, err)
}
