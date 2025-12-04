package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeConfigCmd(t *testing.T) {
	cmd := makeConfigCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "config", cmd.Use)
	assert.True(t, cmd.HasSubCommands())
}

func TestMakeConfigCheckCmd(t *testing.T) {
	cmd := makeConfigCheckCmd()
	assert.NotNil(t, cmd)
	assert.Equal(t, "check", cmd.Use)
	assert.True(t, cmd.Flags().HasFlags())
}

func TestConfigCheckCmd_Help(t *testing.T) {
	cmd := makeConfigCheckCmd()
	cmd.SetArgs([]string{"--help"})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "check")
	assert.Contains(t, output, "Flags:")
	// --profile フラグの確認
	assert.True(t, strings.Contains(output, "--profile") || strings.Contains(output, "-p"))
	// --verbose フラグの確認
	assert.True(t, strings.Contains(output, "--verbose") || strings.Contains(output, "-v"))
}
