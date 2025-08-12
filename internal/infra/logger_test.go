package infra

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	// 色を無効化してテスト環境をシンプルに保つ
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()
	color.NoColor = true

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
				assert.Contains(t, output, "DEBUG debug message")
			} else {
				assert.NotContains(t, output, "DEBUG debug message")
			}
			assert.Contains(t, output, "INFO info message")

			// Test log format and timestamp (RFC3339)
			if len(output) > 0 {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				for _, line := range lines {
					if line == "" {
						continue
					}
					// ログ形式: 時刻 ログレベル ログメッセージ
					parts := strings.SplitN(line, " ", 3)
					assert.GreaterOrEqual(t, len(parts), 3, "Log line should have at least 3 parts: timestamp, level, message")
					if len(parts) >= 3 {
						// タイムスタンプの検証
						_, err := time.Parse(time.RFC3339, parts[0])
						assert.NoError(t, err, "Timestamp should be in RFC3339 format")
						// ログレベルの検証
						assert.Contains(t, []string{"DEBUG", "INFO", "WARN", "ERROR"}, parts[1], "Log level should be valid")
					}
				}
			}
		})
	}
}

func TestSimpleHandler(t *testing.T) {
	// 色を無効化してテスト環境をシンプルに保つ
	originalNoColor := color.NoColor
	defer func() { color.NoColor = originalNoColor }()
	color.NoColor = true

	tests := []struct {
		name          string
		level         slog.Level
		logLevel      slog.Level
		message       string
		expectOutput  bool
		expectPattern string
	}{
		{
			name:          "INFO level message at INFO threshold",
			level:         slog.LevelInfo,
			logLevel:      slog.LevelInfo,
			message:       "test info message",
			expectOutput:  true,
			expectPattern: "INFO test info message",
		},
		{
			name:          "DEBUG level message at INFO threshold",
			level:         slog.LevelInfo,
			logLevel:      slog.LevelDebug,
			message:       "test debug message",
			expectOutput:  false,
			expectPattern: "",
		},
		{
			name:          "ERROR level message at INFO threshold",
			level:         slog.LevelInfo,
			logLevel:      slog.LevelError,
			message:       "test error message",
			expectOutput:  true,
			expectPattern: "ERROR test error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			opts := &slog.HandlerOptions{
				Level: tt.level,
			}
			handler := NewSimpleHandler(&buf, opts)
			logger := slog.New(handler)

			// ログメッセージを出力
			switch tt.logLevel {
			case slog.LevelDebug:
				logger.Debug(tt.message)
			case slog.LevelInfo:
				logger.Info(tt.message)
			case slog.LevelWarn:
				logger.Warn(tt.message)
			case slog.LevelError:
				logger.Error(tt.message)
			}

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, tt.expectPattern)
				// 形式の確認: 時刻 ログレベル メッセージ
				parts := strings.SplitN(strings.TrimSpace(output), " ", 3)
				assert.Len(t, parts, 3)
				// タイムスタンプの検証
				_, err := time.Parse(time.RFC3339, parts[0])
				assert.NoError(t, err)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestSimpleHandler_Handle_WithColors(t *testing.T) {
	// NO_COLOR環境変数を一時的に無効化し、color.NoColorを無効化
	originalEnv := os.Getenv("NO_COLOR")
	originalNoColor := color.NoColor
	defer func() {
		if originalEnv != "" {
			os.Setenv("NO_COLOR", originalEnv)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		color.NoColor = originalNoColor
	}()

	os.Unsetenv("NO_COLOR")
	color.NoColor = false

	tests := []struct {
		name      string
		level     slog.Level
		colorCode string
	}{
		{"DEBUG with gray", slog.LevelDebug, "\\033\\[90m"},
		{"INFO with green", slog.LevelInfo, "\\033\\[32m"},
		{"WARN with yellow", slog.LevelWarn, "\\033\\[33m"},
		{"ERROR with red", slog.LevelError, "\\033\\[31m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := NewSimpleHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

			record := slog.Record{
				Level:   tt.level,
				Message: "test message",
				Time:    time.Now(),
			}

			err := handler.Handle(context.Background(), record)
			assert.NoError(t, err)

			output := buf.String()
			matched, err := regexp.MatchString(tt.colorCode, output)
			assert.NoError(t, err)
			assert.True(t, matched, "Expected color code %s not found in output: %s", tt.colorCode, output)

			// リセットコード（\033[0m）も確認
			resetMatched, err := regexp.MatchString("\\033\\[0m", output)
			assert.NoError(t, err)
			assert.True(t, resetMatched, "Expected reset code \\033[0m not found in output: %s", output)
		})
	}
}

func TestSimpleHandler_Handle_NoColor(t *testing.T) {
	// NO_COLOR環境変数を設定し、color.NoColorを有効化
	originalEnv := os.Getenv("NO_COLOR")
	originalNoColor := color.NoColor
	defer func() {
		if originalEnv != "" {
			os.Setenv("NO_COLOR", originalEnv)
		} else {
			os.Unsetenv("NO_COLOR")
		}
		color.NoColor = originalNoColor
	}()

	os.Setenv("NO_COLOR", "1")
	color.NoColor = true

	var buf bytes.Buffer
	handler := NewSimpleHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})

	record := slog.Record{
		Level:   slog.LevelError,
		Message: "test message",
		Time:    time.Now(),
	}

	err := handler.Handle(context.Background(), record)
	assert.NoError(t, err)

	output := buf.String()
	matched, err := regexp.MatchString("\\033\\[", output)
	assert.NoError(t, err)
	assert.False(t, matched, "ANSI escape codes should not be present when NO_COLOR is set: %s", output)

	// プレーンテキストのログレベルが含まれることを確認
	assert.Contains(t, output, "ERROR test message")
}
