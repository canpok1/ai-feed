package entity

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretString_Value(t *testing.T) {
	tests := []struct {
		name     string
		secret   SecretString
		expected string
	}{
		{
			name:     "正常系_値が正しく返される",
			secret:   SecretString{value: "my-secret-value"},
			expected: "my-secret-value",
		},
		{
			name:     "正常系_空文字列が正しく返される",
			secret:   SecretString{value: ""},
			expected: "",
		},
		{
			name:     "正常系_特殊文字を含む値が正しく返される",
			secret:   SecretString{value: "p@ssw0rd!#$%"},
			expected: "p@ssw0rd!#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.secret.Value()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecretString_String(t *testing.T) {
	tests := []struct {
		name     string
		secret   SecretString
		expected string
	}{
		{
			name:     "正常系_REDACTEDが返される",
			secret:   SecretString{value: "my-secret-value"},
			expected: "[REDACTED]",
		},
		{
			name:     "正常系_空文字列でもREDACTEDが返される",
			secret:   SecretString{value: ""},
			expected: "[REDACTED]",
		},
		{
			name:     "正常系_長い値でもREDACTEDが返される",
			secret:   SecretString{value: "very-long-secret-value-that-should-not-be-shown"},
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.secret.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecretString_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		secret   SecretString
		expected bool
	}{
		{
			name:     "正常系_空文字列の場合にtrueを返す",
			secret:   SecretString{value: ""},
			expected: true,
		},
		{
			name:     "正常系_ゼロ値の場合にtrueを返す",
			secret:   SecretString{},
			expected: true,
		},
		{
			name:     "正常系_値が設定されている場合にfalseを返す",
			secret:   SecretString{value: "my-secret"},
			expected: false,
		},
		{
			name:     "正常系_スペースのみの値の場合にfalseを返す",
			secret:   SecretString{value: "   "},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.secret.IsEmpty()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecretString_LogValue(t *testing.T) {
	tests := []struct {
		name     string
		secret   SecretString
		expected slog.Value
	}{
		{
			name:     "正常系_REDACTEDのslog.Valueが返される",
			secret:   SecretString{value: "my-secret-value"},
			expected: slog.StringValue("[REDACTED]"),
		},
		{
			name:     "正常系_空文字列でもREDACTEDのslog.Valueが返される",
			secret:   SecretString{value: ""},
			expected: slog.StringValue("[REDACTED]"),
		},
		{
			name:     "正常系_機密性の高い値でもREDACTEDのslog.Valueが返される",
			secret:   SecretString{value: "super-secret-password-123"},
			expected: slog.StringValue("[REDACTED]"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.secret.LogValue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecretString_UnmarshalText(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  string
		wantError bool
	}{
		{
			name:      "正常系_バイト列が正しく文字列として設定される",
			input:     []byte("my-secret-value"),
			expected:  "my-secret-value",
			wantError: false,
		},
		{
			name:      "正常系_空のバイト列が空文字列として設定される",
			input:     []byte(""),
			expected:  "",
			wantError: false,
		},
		{
			name:      "正常系_UTF-8文字列が正しく設定される",
			input:     []byte("日本語のシークレット"),
			expected:  "日本語のシークレット",
			wantError: false,
		},
		{
			name:      "正常系_特殊文字を含むバイト列が正しく設定される",
			input:     []byte("p@ssw0rd!#$%^&*()"),
			expected:  "p@ssw0rd!#$%^&*()",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var secret SecretString
			err := secret.UnmarshalText(tt.input)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, secret.Value())
			}
		})
	}
}

func TestSecretString_MarshalText(t *testing.T) {
	tests := []struct {
		name      string
		secret    SecretString
		expected  []byte
		wantError bool
	}{
		{
			name:      "正常系_REDACTEDがバイト列として返される",
			secret:    SecretString{value: "my-secret-value"},
			expected:  []byte("[REDACTED]"),
			wantError: false,
		},
		{
			name:      "正常系_空文字列でもREDACTEDがバイト列として返される",
			secret:    SecretString{value: ""},
			expected:  []byte("[REDACTED]"),
			wantError: false,
		},
		{
			name:      "正常系_どんな値でもREDACTEDがバイト列として返される",
			secret:    SecretString{value: "super-secret-password"},
			expected:  []byte("[REDACTED]"),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.secret.MarshalText()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
