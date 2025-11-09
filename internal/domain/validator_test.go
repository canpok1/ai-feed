package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      ValidationError
		expected ValidationError
	}{
		{
			name: "Required型のエラー",
			err: ValidationError{
				Field:   "ai.gemini.api_key",
				Type:    ValidationErrorTypeRequired,
				Message: "APIキーが設定されていません",
			},
			expected: ValidationError{
				Field:   "ai.gemini.api_key",
				Type:    ValidationErrorTypeRequired,
				Message: "APIキーが設定されていません",
			},
		},
		{
			name: "DummyValue型のエラー",
			err: ValidationError{
				Field:   "ai.gemini.api_key",
				Type:    ValidationErrorTypeDummyValue,
				Message: "ダミー値が設定されています",
			},
			expected: ValidationError{
				Field:   "ai.gemini.api_key",
				Type:    ValidationErrorTypeDummyValue,
				Message: "ダミー値が設定されています",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.Field, tt.err.Field)
			assert.Equal(t, tt.expected.Type, tt.err.Type)
			assert.Equal(t, tt.expected.Message, tt.err.Message)
		})
	}
}

func TestValidationResult(t *testing.T) {
	tests := []struct {
		name     string
		result   ValidationResult
		expected ValidationResult
	}{
		{
			name: "成功のケース",
			result: ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
				Summary: ConfigSummary{
					GeminiConfigured:        true,
					GeminiModel:             "gemini-1.5-flash",
					CommentPromptConfigured: true,
					SlackConfigured:         true,
					MisskeyConfigured:       false,
				},
			},
			expected: ValidationResult{
				Valid:  true,
				Errors: []ValidationError{},
				Summary: ConfigSummary{
					GeminiConfigured:        true,
					GeminiModel:             "gemini-1.5-flash",
					CommentPromptConfigured: true,
					SlackConfigured:         true,
					MisskeyConfigured:       false,
				},
			},
		},
		{
			name: "失敗のケース",
			result: ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{
						Field:   "ai.gemini.api_key",
						Type:    ValidationErrorTypeRequired,
						Message: "APIキーが設定されていません",
					},
				},
				Summary: ConfigSummary{
					GeminiConfigured:        false,
					GeminiModel:             "",
					CommentPromptConfigured: false,
					SlackConfigured:         false,
					MisskeyConfigured:       false,
				},
			},
			expected: ValidationResult{
				Valid: false,
				Errors: []ValidationError{
					{
						Field:   "ai.gemini.api_key",
						Type:    ValidationErrorTypeRequired,
						Message: "APIキーが設定されていません",
					},
				},
				Summary: ConfigSummary{
					GeminiConfigured:        false,
					GeminiModel:             "",
					CommentPromptConfigured: false,
					SlackConfigured:         false,
					MisskeyConfigured:       false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.Valid, tt.result.Valid)
			assert.Equal(t, tt.expected.Errors, tt.result.Errors)
			assert.Equal(t, tt.expected.Summary, tt.result.Summary)
		})
	}
}

func TestConfigSummary(t *testing.T) {
	tests := []struct {
		name     string
		summary  ConfigSummary
		expected ConfigSummary
	}{
		{
			name: "全て設定済みのケース",
			summary: ConfigSummary{
				GeminiConfigured:        true,
				GeminiModel:             "gemini-1.5-flash",
				CommentPromptConfigured: true,
				SlackConfigured:         true,
				MisskeyConfigured:       true,
			},
			expected: ConfigSummary{
				GeminiConfigured:        true,
				GeminiModel:             "gemini-1.5-flash",
				CommentPromptConfigured: true,
				SlackConfigured:         true,
				MisskeyConfigured:       true,
			},
		},
		{
			name: "一部未設定のケース",
			summary: ConfigSummary{
				GeminiConfigured:        true,
				GeminiModel:             "gemini-1.5-flash",
				CommentPromptConfigured: true,
				SlackConfigured:         false,
				MisskeyConfigured:       false,
			},
			expected: ConfigSummary{
				GeminiConfigured:        true,
				GeminiModel:             "gemini-1.5-flash",
				CommentPromptConfigured: true,
				SlackConfigured:         false,
				MisskeyConfigured:       false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.summary)
		})
	}
}
