package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeminiConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *GeminiConfig
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_TypeとAPIKeyが適切に設定されている",
			config: &GeminiConfig{
				Type:   "gemini-pro",
				APIKey: "valid-api-key",
			},
			wantErr: false,
		},
		{
			name: "異常系_Typeが空文字列",
			config: &GeminiConfig{
				Type:   "",
				APIKey: "valid-api-key",
			},
			wantErr: true,
			errors:  []string{"Gemini設定のTypeが設定されていません"},
		},
		{
			name: "異常系_APIKeyが空文字列",
			config: &GeminiConfig{
				Type:   "gemini-pro",
				APIKey: "",
			},
			wantErr: true,
			errors:  []string{"Gemini API keyが設定されていません"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestProfile_Validate(t *testing.T) {
	validAI := &AIConfig{
		Gemini: &GeminiConfig{
			Type:   "gemini-pro",
			APIKey: "valid-api-key",
		},
	}
	validPrompt := &PromptConfig{
		SystemPrompt:          "システムプロンプト",
		CommentPromptTemplate: "コメントテンプレート",
	}
	validOutput := &OutputConfig{
		SlackAPI: &SlackAPIConfig{
			APIToken: "valid-token",
			Channel:  "#general",
		},
	}

	tests := []struct {
		name    string
		profile *Profile
		wantErr bool
	}{
		{
			name: "正常系_すべての設定が適切",
			profile: &Profile{
				AI:     validAI,
				Prompt: validPrompt,
				Output: validOutput,
			},
			wantErr: false,
		},
		{
			name: "異常系_AI設定がnil",
			profile: &Profile{
				AI:     nil,
				Prompt: validPrompt,
				Output: validOutput,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.Validate()
			assert.Equal(t, !tt.wantErr, result.IsValid)
		})
	}
}
