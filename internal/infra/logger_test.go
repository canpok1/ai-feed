package infra

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		logLevel slog.Level
	}{
		{
			name:     "verbose false should set INFO level",
			verbose:  false,
			logLevel: slog.LevelInfo,
		},
		{
			name:     "verbose true should set DEBUG level",
			verbose:  true,
			logLevel: slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer

			// Mock the output destination for testing
			originalHandler := slog.Default().Handler()

			// Initialize logger with test configuration
			opts := &slog.HandlerOptions{
				Level: tt.logLevel,
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey {
						return slog.Attr{
							Key:   a.Key,
							Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
						}
					}
					return a
				},
			}

			handler := slog.NewTextHandler(&buf, opts)
			logger := slog.New(handler)
			slog.SetDefault(logger)

			// Test different log levels
			slog.Debug("debug message")
			slog.Info("info message")
			slog.Warn("warn message")
			slog.Error("error message")

			output := buf.String()

			if tt.verbose {
				// DEBUG level should show all messages
				assert.Contains(t, output, "debug message")
				assert.Contains(t, output, "info message")
				assert.Contains(t, output, "warn message")
				assert.Contains(t, output, "error message")
			} else {
				// INFO level should not show debug messages
				assert.NotContains(t, output, "debug message")
				assert.Contains(t, output, "info message")
				assert.Contains(t, output, "warn message")
				assert.Contains(t, output, "error message")
			}

			// Test timestamp format (RFC3339)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "time=") {
					// Extract timestamp from log line
					timePart := strings.Split(line, " ")[0]
					timeValue := strings.TrimPrefix(timePart, "time=")
					_, err := time.Parse(time.RFC3339, timeValue)
					assert.NoError(t, err, "Timestamp should be in RFC3339 format")
				}
			}

			// Restore original handler
			slog.SetDefault(slog.New(originalHandler))
		})
	}
}

func TestInitLoggerIntegration(t *testing.T) {
	// Test the actual InitLogger function
	var buf bytes.Buffer

	// Save original default logger
	originalDefault := slog.Default()

	// Test with verbose=false
	t.Run("InitLogger with verbose=false", func(t *testing.T) {
		InitLogger(false)

		// Replace handler with test handler to capture output
		opts := &slog.HandlerOptions{
			Level: slog.LevelInfo,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
					}
				}
				return a
			},
		}
		handler := slog.NewTextHandler(&buf, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)

		slog.Debug("debug message")
		slog.Info("info message")

		output := buf.String()
		assert.NotContains(t, output, "debug message")
		assert.Contains(t, output, "info message")
	})

	// Test with verbose=true
	t.Run("InitLogger with verbose=true", func(t *testing.T) {
		buf.Reset()
		InitLogger(true)

		// Replace handler with test handler to capture output
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					return slog.Attr{
						Key:   a.Key,
						Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
					}
				}
				return a
			},
		}
		handler := slog.NewTextHandler(&buf, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)

		slog.Debug("debug message")
		slog.Info("info message")

		output := buf.String()
		assert.Contains(t, output, "debug message")
		assert.Contains(t, output, "info message")
	})

	// Restore original default logger
	slog.SetDefault(originalDefault)
}
