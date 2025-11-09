package infra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/infra/profile"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidator_Validate_Success(t *testing.T) {
	tests := []struct {
		name    string
		config  *infra.Config
		profile *entity.Profile
		want    *domain.ValidationResult
	}{
		{
			name: "すべての必須項目が正しく設定されている",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
			},
			want: &domain.ValidationResult{
				Valid:  true,
				Errors: []domain.ValidationError{},
				Summary: domain.ConfigSummary{
					GeminiConfigured:        true,
					GeminiModel:             "gemini-1.5-flash",
					CommentPromptConfigured: true,
					SlackConfigured:         false,
					MisskeyConfigured:       false,
				},
			},
		},
		{
			name: "Slack APIが有効で正しく設定されている",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
				Output: &entity.OutputConfig{
					SlackAPI: &entity.SlackAPIConfig{
						Enabled:  true,
						APIToken: entity.NewSecretString("valid-slack-token"),
						Channel:  "test-channel",
					},
				},
			},
			want: &domain.ValidationResult{
				Valid:  true,
				Errors: []domain.ValidationError{},
				Summary: domain.ConfigSummary{
					GeminiConfigured:        true,
					GeminiModel:             "gemini-1.5-flash",
					CommentPromptConfigured: true,
					SlackConfigured:         true,
					MisskeyConfigured:       false,
				},
			},
		},
		{
			name: "Misskeyが有効で正しく設定されている",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
				Output: &entity.OutputConfig{
					Misskey: &entity.MisskeyConfig{
						Enabled:  true,
						APIToken: entity.NewSecretString("valid-misskey-token"),
						APIURL:   "https://misskey.example.com",
					},
				},
			},
			want: &domain.ValidationResult{
				Valid:  true,
				Errors: []domain.ValidationError{},
				Summary: domain.ConfigSummary{
					GeminiConfigured:        true,
					GeminiModel:             "gemini-1.5-flash",
					CommentPromptConfigured: true,
					SlackConfigured:         false,
					MisskeyConfigured:       true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := infra.NewConfigValidator(tt.config, tt.profile)
			result, err := validator.Validate()
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Valid, result.Valid)
			assert.Equal(t, tt.want.Errors, result.Errors)
			assert.Equal(t, tt.want.Summary, result.Summary)
		})
	}
}

func TestConfigValidator_Validate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		config      *infra.Config
		profile     *entity.Profile
		expectValid bool
		expectError []domain.ValidationError
	}{
		{
			name: "AI設定が未設定",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
			},
			expectValid: false,
			expectError: []domain.ValidationError{
				{
					Field:   "ai",
					Type:    domain.ValidationErrorTypeRequired,
					Message: "AI設定が設定されていません",
				},
			},
		},
		{
			name: "Gemini APIKeyがダミー値",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("xxxxxx"),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
			},
			expectValid: false,
			expectError: []domain.ValidationError{
				{
					Field:   "ai.gemini.api_key",
					Type:    domain.ValidationErrorTypeDummyValue,
					Message: "Gemini APIキーがダミー値です: \"xxxxxx\"",
				},
			},
		},
		{
			name: "プロンプト設定が未設定",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
			},
			expectValid: false,
			expectError: []domain.ValidationError{
				{
					Field:   "prompt",
					Type:    domain.ValidationErrorTypeRequired,
					Message: "プロンプト設定が設定されていません",
				},
			},
		},
		{
			name: "コメントプロンプトテンプレートが未設定",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
				Prompt: &entity.PromptConfig{},
			},
			expectValid: false,
			expectError: []domain.ValidationError{
				{
					Field:   "comment_prompt_template",
					Type:    domain.ValidationErrorTypeRequired,
					Message: "コメントプロンプトテンプレートが設定されていません",
				},
			},
		},
		{
			name: "Misskey APIトークンがダミー値",
			config: &infra.Config{
				DefaultProfile: &infra.Profile{},
			},
			profile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString("valid-api-key-12345"),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
				Output: &entity.OutputConfig{
					Misskey: &entity.MisskeyConfig{
						Enabled:  true,
						APIToken: entity.NewSecretString("YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE"),
						APIURL:   "https://misskey.example.com",
					},
				},
			},
			expectValid: false,
			expectError: []domain.ValidationError{
				{
					Field:   "output.misskey.api_token",
					Type:    domain.ValidationErrorTypeDummyValue,
					Message: "Misskey APIトークンがダミー値です: \"YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE\"",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := infra.NewConfigValidator(tt.config, tt.profile)
			result, err := validator.Validate()
			assert.NoError(t, err)
			assert.Equal(t, tt.expectValid, result.Valid)
			assert.Equal(t, tt.expectError, result.Errors)
		})
	}
}

func TestConfigValidator_WithProfileMerge(t *testing.T) {
	// 実際のファイルを使ったテスト
	tmpDir := t.TempDir()

	// config.ymlを作成
	configPath := filepath.Join(tmpDir, "config.yml")
	configRepo := infra.NewYamlConfigRepository(configPath)
	config := &infra.Config{
		DefaultProfile: &infra.Profile{
			AI: &infra.AIConfig{
				Gemini: &infra.GeminiConfig{
					Type:   "gemini-1.5-flash",
					APIKey: "xxxxxx", // ダミー値
				},
			},
			Prompt: &infra.PromptConfig{
				CommentPromptTemplate: "default template",
			},
		},
	}
	err := configRepo.Save(config)
	assert.NoError(t, err)

	// profile.ymlを作成
	profilePath := filepath.Join(tmpDir, "profile.yml")
	profileRepo := profile.NewYamlProfileRepositoryImpl(profilePath)

	// プロファイルファイルを手動で作成（テストのため）
	profileFile, err := os.Create(profilePath)
	assert.NoError(t, err)
	defer profileFile.Close()

	profileContent := `ai:
  gemini:
    api_key: valid-profile-api-key
`
	_, err = profileFile.WriteString(profileContent)
	assert.NoError(t, err)
	profileFile.Close()

	// configを読み込み
	loadedConfig, err := configRepo.Load()
	assert.NoError(t, err)

	// profileを読み込み
	loadedProfile, err := profileRepo.LoadProfile()
	assert.NoError(t, err)

	// default_profileをentity.Profileに変換
	baseProfile, err := loadedConfig.DefaultProfile.ToEntity()
	assert.NoError(t, err)

	// profileをマージ
	baseProfile.Merge(loadedProfile)

	// バリデーション
	validator := infra.NewConfigValidator(loadedConfig, baseProfile)
	result, err := validator.Validate()
	assert.NoError(t, err)
	assert.True(t, result.Valid, "profileをマージした結果、バリデーションが成功するべき")
	assert.Empty(t, result.Errors)
}

func TestIsDummyValue(t *testing.T) {
	// このテストは内部関数なので、直接テストできない
	// 代わりにバリデーション経由で間接的にテストする
	tests := []struct {
		name        string
		apiKey      string
		expectError bool
	}{
		{
			name:        "xxxxxx はダミー値",
			apiKey:      "xxxxxx",
			expectError: true,
		},
		{
			name:        "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE はダミー値",
			apiKey:      "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
			expectError: true,
		},
		{
			name:        "valid-api-key-12345 はダミー値ではない",
			apiKey:      "valid-api-key-12345",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &infra.Config{
				DefaultProfile: &infra.Profile{},
			}
			profile := &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   "gemini-1.5-flash",
						APIKey: entity.NewSecretString(tt.apiKey),
					},
				},
				Prompt: &entity.PromptConfig{
					CommentPromptTemplate: "test prompt template",
				},
			}

			validator := infra.NewConfigValidator(config, profile)
			result, err := validator.Validate()
			assert.NoError(t, err)

			if tt.expectError {
				assert.False(t, result.Valid)
				assert.NotEmpty(t, result.Errors)
			} else {
				assert.True(t, result.Valid)
				assert.Empty(t, result.Errors)
			}
		})
	}
}
