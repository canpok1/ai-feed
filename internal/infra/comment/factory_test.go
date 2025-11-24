package comment

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestCommentGeneratorFactory_MakeCommentGenerator(t *testing.T) {
	tests := []struct {
		name      string
		model     *entity.AIConfig
		prompt    *entity.PromptConfig
		wantErr   bool
		errString string
	}{
		{
			name:      "異常系_modelがnil",
			model:     nil,
			prompt:    &entity.PromptConfig{SystemPrompt: "test"},
			wantErr:   true,
			errString: "model is nil",
		},
		{
			name: "異常系_promptがnil",
			model: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type:   "gemini-2.5-flash",
					APIKey: entity.NewSecretString("test-key"),
				},
			},
			prompt:    nil,
			wantErr:   true,
			errString: "prompt is nil",
		},
		{
			name: "異常系_Gemini設定がnil",
			model: &entity.AIConfig{
				Gemini: nil,
			},
			prompt:    &entity.PromptConfig{SystemPrompt: "test"},
			wantErr:   true,
			errString: "gemini config is nil",
		},
		{
			name: "異常系_Geminiモデルタイプが空文字列",
			model: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type:   "",
					APIKey: entity.NewSecretString("test-key"),
				},
			},
			prompt:    &entity.PromptConfig{SystemPrompt: "test"},
			wantErr:   true,
			errString: "gemini model type is empty",
		},
		{
			name: "正常系_任意のGeminiモデル名_gemini-1.5-pro",
			model: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type:   "gemini-1.5-pro",
					APIKey: entity.NewSecretString("test-key"),
				},
			},
			prompt: &entity.PromptConfig{
				SystemPrompt:          "test system prompt",
				CommentPromptTemplate: "test template",
			},
			wantErr: false,
		},
		{
			name: "正常系_任意のGeminiモデル名_gemini-2.0-flash-exp",
			model: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type:   "gemini-2.0-flash-exp",
					APIKey: entity.NewSecretString("test-key"),
				},
			},
			prompt: &entity.PromptConfig{
				SystemPrompt:          "test system prompt",
				CommentPromptTemplate: "test template",
			},
			wantErr: false,
		},
		{
			name: "正常系_任意のGeminiモデル名_gemini-2.5-flash",
			model: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type:   "gemini-2.5-flash",
					APIKey: entity.NewSecretString("test-key"),
				},
			},
			prompt: &entity.PromptConfig{
				SystemPrompt:          "test system prompt",
				CommentPromptTemplate: "test template",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewCommentGeneratorFactory()
			generator, err := factory.MakeCommentGenerator(tt.model, tt.prompt)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, generator)
				if tt.errString != "" {
					assert.Contains(t, err.Error(), tt.errString)
				}
			} else {
				// 注意: 正常系では実際のGeminiクライアント作成を試みるため、
				// APIキーが無効な場合はエラーになる可能性がある
				// ここでは、バリデーションを通過することのみを確認
				// （実際のクライアント作成の成功/失敗はこのテストのスコープ外）
				if err != nil {
					// Geminiクライアントの作成エラーは許容（APIキーが無効なため）
					assert.Contains(t, err.Error(), "failed to create gemini client")
				} else {
					assert.NotNil(t, generator)
				}
			}
		})
	}
}
