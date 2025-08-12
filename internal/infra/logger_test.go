package infra

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	// 元のos.Stdoutとslogのデフォルトを保存
	originalStdout := os.Stdout
	originalLogger := slog.Default()
	defer func() {
		os.Stdout = originalStdout
		slog.SetDefault(originalLogger)
	}()

	tests := []struct {
		name        string
		verbose     bool
		expectDebug bool
	}{
		{
			name:        "verbose false should set INFO level",
			verbose:     false,
			expectDebug: false,
		},
		{
			name:        "verbose true should set DEBUG level",
			verbose:     true,
			expectDebug: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, w, err := os.Pipe()
			assert.NoError(t, err)
			os.Stdout = w

			InitLogger(tt.verbose)

			slog.Debug("debug message")
			slog.Info("info message")

			w.Close()
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			assert.NoError(t, err)

			output := buf.String()

			if tt.expectDebug {
				assert.Contains(t, output, "debug message")
			} else {
				assert.NotContains(t, output, "debug message")
			}
			assert.Contains(t, output, "info message")

			// Test timestamp format (RFC3339)
			if len(output) > 0 {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) > 0 && lines[0] != "" {
					timePart := strings.Split(lines[0], " ")[0]
					timeValue := strings.TrimPrefix(timePart, "time=")
					_, err := time.Parse(time.RFC3339, timeValue)
					assert.NoError(t, err, "Timestamp should be in RFC3339 format")
				}
			}
		})
	}
}
