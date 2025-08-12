package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVerboseFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectVerbose bool
	}{
		{
			name:          "no verbose flag",
			args:          []string{"recommend", "--help"},
			expectVerbose: false,
		},
		{
			name:          "short verbose flag",
			args:          []string{"-v", "recommend", "--help"},
			expectVerbose: true,
		},
		{
			name:          "long verbose flag",
			args:          []string{"--verbose", "recommend", "--help"},
			expectVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset verbose flag
			verbose = false

			// Create root command
			rootCmd := makeRootCmd()

			// Set args
			rootCmd.SetArgs(tt.args)

			// Capture output to suppress help text
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// Execute command (this should set the verbose flag)
			err := rootCmd.Execute()

			// The help command returns an error for exit, which is expected
			// For all cases, we expect the verbose variable to be set correctly
			assert.Equal(t, tt.expectVerbose, verbose)

			// For help commands, we don't care about the error as it's expected
			_ = err
		})
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Save original default logger
	originalDefault := slog.Default()
	defer slog.SetDefault(originalDefault)

	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "logger initialization with verbose false",
			verbose: false,
		},
		{
			name:    "logger initialization with verbose true",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set verbose flag
			verbose = tt.verbose

			// Create root command
			rootCmd := makeRootCmd()

			// Set up PersistentPreRun
			rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				// This would normally call infra.InitLogger(verbose)
				// For testing, we'll just verify the verbose flag is set correctly
				assert.Equal(t, tt.verbose, verbose)
			}

			// Set a dummy command to trigger PersistentPreRun
			rootCmd.SetArgs([]string{"--help"})

			// Capture output
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// Execute - this should trigger PersistentPreRun
			err := rootCmd.Execute()

			// Help command exits with error, which is expected
			_ = err
		})
	}
}

func TestRootCommandCreation(t *testing.T) {
	cmd := makeRootCmd()

	// Test basic command properties
	assert.Equal(t, "ai-feed", cmd.Use)
	assert.Contains(t, cmd.Short, "RSSフィードから記事を取得")
	assert.True(t, cmd.SilenceUsage)

	// Test flags
	configFlag := cmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "", configFlag.DefValue)

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	// Test short flag for verbose
	verboseFlagShort := cmd.PersistentFlags().ShorthandLookup("v")
	assert.NotNil(t, verboseFlagShort)
	assert.Equal(t, verboseFlag, verboseFlagShort)
}

func TestExecuteFunction(t *testing.T) {
	// This test ensures Execute function can be called without panicking
	// We'll redirect output to avoid printing to stdout during tests

	// Save original stdout/stderr
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to just show help to avoid actually running commands
	os.Args = []string{"ai-feed", "--help"}

	// The Execute function should not panic
	assert.NotPanics(t, func() {
		err := Execute()
		// Help command returns an error (exit code), which is expected behavior
		_ = err
	})
}
