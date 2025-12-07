//go:build integration

// Package config はプロファイルマージロジックの統合テストを提供する
package config

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProfileMerge_BasicMerge はデフォルトとファイル設定の正確なマージ動作を確認する
func TestProfileMerge_BasicMerge(t *testing.T) {
	// デフォルトプロファイルを作成
	defaultProfile := ValidEntityProfile()

	// ファイルからのプロファイル（部分的な設定）
	fileProfile := &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{
				Type: "gemini-2.0-flash", // デフォルト値を上書き
			},
		},
	}

	// マージ実行
	defaultProfile.Merge(fileProfile)

	// 結果の検証
	require.NotNil(t, defaultProfile.AI)
	require.NotNil(t, defaultProfile.AI.Gemini)
	assert.Equal(t, "gemini-2.0-flash", defaultProfile.AI.Gemini.Type,
		"ファイル設定でTypeが上書きされるはずです")
	assert.Equal(t, "test-api-key", defaultProfile.AI.Gemini.APIKey.Value(),
		"APIKeyはデフォルト値が維持されるはずです")
}

// TestProfileMerge_FileOverridesDefault はファイル設定によるデフォルト値の上書きを検証する
func TestProfileMerge_FileOverridesDefault(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue string
		fileValue    string
		expected     string
	}{
		{
			name:         "ファイル設定がデフォルト値を上書き",
			defaultValue: "default-value",
			fileValue:    "file-value",
			expected:     "file-value",
		},
		{
			name:         "ファイル設定が空の場合はデフォルト値を維持",
			defaultValue: "default-value",
			fileValue:    "",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultProfile := &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type:   tt.defaultValue,
						APIKey: entity.NewSecretString("api-key"),
					},
				},
				Prompt: NewEntityPromptConfig(),
				Output: NewEntityOutputConfig(),
			}

			fileProfile := &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{
						Type: tt.fileValue,
					},
				},
			}

			defaultProfile.Merge(fileProfile)

			assert.Equal(t, tt.expected, defaultProfile.AI.Gemini.Type)
		})
	}
}

// TestProfileMerge_NestedStructure はネスト構造のマージを検証する
func TestProfileMerge_NestedStructure(t *testing.T) {
	t.Run("AI.Geminiのネストマージ", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalAPIKey := defaultProfile.AI.Gemini.APIKey.Value()

		// Type のみを上書きするプロファイル
		fileProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type: "gemini-1.5-pro",
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, "gemini-1.5-pro", defaultProfile.AI.Gemini.Type,
			"Gemini.Typeが上書きされるはずです")
		assert.Equal(t, originalAPIKey, defaultProfile.AI.Gemini.APIKey.Value(),
			"Gemini.APIKeyはデフォルト値が維持されるはずです")
	})

	t.Run("Output.SlackAPIのネストマージ", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalChannel := defaultProfile.Output.SlackAPI.Channel

		// チャンネルのみを上書きするプロファイル
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Channel: "#new-channel",
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, "#new-channel", defaultProfile.Output.SlackAPI.Channel,
			"SlackAPI.Channelが上書きされるはずです")
		assert.NotEqual(t, originalChannel, defaultProfile.Output.SlackAPI.Channel,
			"チャンネルが変更されているはずです")
	})

	t.Run("Output.Misskeyのネストマージ", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalAPIURL := defaultProfile.Output.Misskey.APIURL

		// API URLのみを上書きするプロファイル
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					APIURL: "http://new-misskey.example.com",
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, "http://new-misskey.example.com", defaultProfile.Output.Misskey.APIURL,
			"Misskey.APIURLが上書きされるはずです")
		assert.NotEqual(t, originalAPIURL, defaultProfile.Output.Misskey.APIURL,
			"API URLが変更されているはずです")
	})
}

// TestProfileMerge_PreserveOmittedFields は省略項目のデフォルト値維持を確認する
func TestProfileMerge_PreserveOmittedFields(t *testing.T) {
	defaultProfile := ValidEntityProfile()
	originalPrompt := defaultProfile.Prompt.SystemPrompt
	originalAPIKey := defaultProfile.AI.Gemini.APIKey.Value()
	originalSlackToken := defaultProfile.Output.SlackAPI.APIToken.Value()

	// 一部の設定のみを含むプロファイル
	fileProfile := &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{
				Type: "gemini-2.0-flash",
				// APIKeyは省略
			},
		},
		// Prompt、Outputは省略
	}

	defaultProfile.Merge(fileProfile)

	// 省略された項目がデフォルト値を維持していることを確認
	assert.Equal(t, originalPrompt, defaultProfile.Prompt.SystemPrompt,
		"省略されたPrompt設定はデフォルト値を維持するはずです")
	assert.Equal(t, originalAPIKey, defaultProfile.AI.Gemini.APIKey.Value(),
		"省略されたAPIKeyはデフォルト値を維持するはずです")
	assert.Equal(t, originalSlackToken, defaultProfile.Output.SlackAPI.APIToken.Value(),
		"省略されたSlackトークンはデフォルト値を維持するはずです")
}

// TestProfileMerge_ValidationAfterMerge はマージ後のバリデーション動作を確認する
func TestProfileMerge_ValidationAfterMerge(t *testing.T) {
	t.Run("正常なマージ後のバリデーション成功", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()

		// 有効な値で上書き
		fileProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type: "gemini-2.0-flash",
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		result := defaultProfile.Validate()
		assert.True(t, result.IsValid, "マージ後のバリデーションは成功するはずです")
		assert.Empty(t, result.Errors, "エラーがないはずです")
	})

	t.Run("不正値でのマージ後のバリデーション失敗", func(t *testing.T) {
		// 最小限のデフォルトプロファイル（APIKeyなし）
		defaultProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type: "gemini-2.5-flash",
					// APIKeyなし
				},
			},
			Prompt: NewEntityPromptConfig(),
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled: false, // 無効化されているのでバリデーションスキップ
				},
				Misskey: &entity.MisskeyConfig{
					Enabled: false, // 無効化されているのでバリデーションスキップ
				},
			},
		}

		// ファイルプロファイルもAPIKeyを提供しない
		fileProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					Type: "gemini-2.0-flash",
					// APIKeyなし
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		result := defaultProfile.Validate()
		assert.False(t, result.IsValid, "APIKeyがない場合、バリデーションは失敗するはずです")
		assert.Contains(t, result.Errors, "Gemini APIキーが設定されていません",
			"APIKeyに関するエラーメッセージが含まれるはずです")
	})
}

// TestProfileMerge_PartialRequiredFieldCompletion は部分的な必須項目補完後の正常動作を確認する
func TestProfileMerge_PartialRequiredFieldCompletion(t *testing.T) {
	// デフォルトにAPIKeyのみ設定
	defaultProfile := &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{
				APIKey: entity.NewSecretString("default-api-key"),
				// Typeなし
			},
		},
		Prompt: NewEntityPromptConfig(),
		Output: &entity.OutputConfig{
			SlackAPI: &entity.SlackAPIConfig{Enabled: false},
			Misskey:  &entity.MisskeyConfig{Enabled: false},
		},
	}

	// ファイルでTypeを補完
	fileProfile := &entity.Profile{
		AI: &entity.AIConfig{
			Gemini: &entity.GeminiConfig{
				Type: "gemini-2.5-flash",
				// APIKeyなし（デフォルトから継承）
			},
		},
	}

	defaultProfile.Merge(fileProfile)

	// 両方の必須項目が揃っていることを確認
	require.NotNil(t, defaultProfile.AI.Gemini)
	assert.Equal(t, "gemini-2.5-flash", defaultProfile.AI.Gemini.Type,
		"Typeがファイルから補完されるはずです")
	assert.Equal(t, "default-api-key", defaultProfile.AI.Gemini.APIKey.Value(),
		"APIKeyがデフォルトから維持されるはずです")

	// バリデーション成功を確認
	result := defaultProfile.Validate()
	assert.True(t, result.IsValid, "両方の必須項目が揃えばバリデーションは成功するはずです")
}

// TestProfileMerge_MultiLevelNestedOverride は複数レベルネストでの部分上書きを検証する
func TestProfileMerge_MultiLevelNestedOverride(t *testing.T) {
	t.Run("ai.gemini.api_keyのみ上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalType := defaultProfile.AI.Gemini.Type

		// APIKeyのみを上書き
		fileProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Gemini: &entity.GeminiConfig{
					APIKey: entity.NewSecretString("new-api-key"),
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, originalType, defaultProfile.AI.Gemini.Type,
			"Typeはデフォルト値が維持されるはずです")
		assert.Equal(t, "new-api-key", defaultProfile.AI.Gemini.APIKey.Value(),
			"APIKeyが新しい値で上書きされるはずです")
	})

	t.Run("output.slack_api.channelのみ上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalToken := defaultProfile.Output.SlackAPI.APIToken.Value()
		originalEnabled := defaultProfile.Output.SlackAPI.Enabled

		// チャンネルのみを上書き
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Channel: "#override-channel",
					Enabled: originalEnabled, // Enabledはboolなのでゼロ値との区別が困難、明示的に設定
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, "#override-channel", defaultProfile.Output.SlackAPI.Channel,
			"Channelが上書きされるはずです")
		assert.Equal(t, originalToken, defaultProfile.Output.SlackAPI.APIToken.Value(),
			"APITokenはデフォルト値が維持されるはずです")
	})

	t.Run("prompt設定の部分上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalSystemPrompt := defaultProfile.Prompt.SystemPrompt
		originalSelectorPrompt := defaultProfile.Prompt.SelectorPrompt

		// CommentPromptTemplateのみを上書き
		fileProfile := &entity.Profile{
			Prompt: &entity.PromptConfig{
				CommentPromptTemplate: "新しいテンプレート: {{.Title}}",
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.Equal(t, originalSystemPrompt, defaultProfile.Prompt.SystemPrompt,
			"SystemPromptはデフォルト値が維持されるはずです")
		assert.Equal(t, "新しいテンプレート: {{.Title}}", defaultProfile.Prompt.CommentPromptTemplate,
			"CommentPromptTemplateが上書きされるはずです")
		assert.Equal(t, originalSelectorPrompt, defaultProfile.Prompt.SelectorPrompt,
			"SelectorPromptはデフォルト値が維持されるはずです")
	})
}

// TestProfileMerge_NilHandling はnilプロファイルのマージ処理を検証する
func TestProfileMerge_NilHandling(t *testing.T) {
	t.Run("nilプロファイルをマージしても変更なし", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalType := defaultProfile.AI.Gemini.Type
		originalAPIKey := defaultProfile.AI.Gemini.APIKey.Value()

		defaultProfile.Merge(nil)

		assert.Equal(t, originalType, defaultProfile.AI.Gemini.Type,
			"nilマージ後もTypeは変更されないはずです")
		assert.Equal(t, originalAPIKey, defaultProfile.AI.Gemini.APIKey.Value(),
			"nilマージ後もAPIKeyは変更されないはずです")
	})

	t.Run("空のプロファイルをマージしても変更なし", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		originalType := defaultProfile.AI.Gemini.Type

		emptyProfile := &entity.Profile{}

		defaultProfile.Merge(emptyProfile)

		assert.Equal(t, originalType, defaultProfile.AI.Gemini.Type,
			"空プロファイルマージ後もTypeは変更されないはずです")
	})
}

// TestProfileMerge_BooleanFieldHandling はboolフィールドのマージ処理を検証する
func TestProfileMerge_BooleanFieldHandling(t *testing.T) {
	t.Run("Enabled=falseで上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()
		assert.True(t, defaultProfile.Output.SlackAPI.Enabled,
			"デフォルトはEnabledがtrueのはずです")

		// Enabled=falseで上書き
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled: false,
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		// boolフィールドは常に上書きされる
		assert.False(t, defaultProfile.Output.SlackAPI.Enabled,
			"Enabledがfalseに上書きされるはずです")
	})

	t.Run("Enabled=trueで上書き", func(t *testing.T) {
		defaultProfile := &entity.Profile{
			AI:     NewEntityAIConfig(),
			Prompt: NewEntityPromptConfig(),
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled:  false,
					APIToken: entity.NewSecretString("token"),
					Channel:  "#channel",
				},
			},
		}

		// Enabled=trueで上書き
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					Enabled: true,
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		assert.True(t, defaultProfile.Output.SlackAPI.Enabled,
			"Enabledがtrueに上書きされるはずです")
	})
}

// TestProfileMerge_MessageTemplateOverride はMessageTemplateの上書きを検証する
func TestProfileMerge_MessageTemplateOverride(t *testing.T) {
	t.Run("SlackAPIのMessageTemplate上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()

		newTemplate := "新しいSlackテンプレート: {{.Article.Title}}"
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{
					MessageTemplate: &newTemplate,
					Enabled:         true,
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		require.NotNil(t, defaultProfile.Output.SlackAPI.MessageTemplate)
		assert.Equal(t, newTemplate, *defaultProfile.Output.SlackAPI.MessageTemplate,
			"MessageTemplateが上書きされるはずです")
	})

	t.Run("MisskeyのMessageTemplate上書き", func(t *testing.T) {
		defaultProfile := ValidEntityProfile()

		newTemplate := "新しいMisskeyテンプレート: {{.Article.Title}}"
		fileProfile := &entity.Profile{
			Output: &entity.OutputConfig{
				Misskey: &entity.MisskeyConfig{
					MessageTemplate: &newTemplate,
					Enabled:         true,
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		require.NotNil(t, defaultProfile.Output.Misskey.MessageTemplate)
		assert.Equal(t, newTemplate, *defaultProfile.Output.Misskey.MessageTemplate,
			"MessageTemplateが上書きされるはずです")
	})
}

// TestProfileMerge_MockConfigMerge はMock設定のマージを検証する
func TestProfileMerge_MockConfigMerge(t *testing.T) {
	t.Run("Mock設定のマージ", func(t *testing.T) {
		defaultProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Mock: &entity.MockConfig{
					Enabled:      false,
					SelectorMode: "first",
					Comment:      "デフォルトコメント",
				},
			},
			Prompt: NewEntityPromptConfig(),
			Output: &entity.OutputConfig{
				SlackAPI: &entity.SlackAPIConfig{Enabled: false},
				Misskey:  &entity.MisskeyConfig{Enabled: false},
			},
		}

		fileProfile := &entity.Profile{
			AI: &entity.AIConfig{
				Mock: &entity.MockConfig{
					Enabled:      true,
					SelectorMode: "random",
					// Commentは省略
				},
			},
		}

		defaultProfile.Merge(fileProfile)

		require.NotNil(t, defaultProfile.AI.Mock)
		assert.True(t, defaultProfile.AI.Mock.Enabled,
			"Enabledがtrueに上書きされるはずです")
		assert.Equal(t, "random", defaultProfile.AI.Mock.SelectorMode,
			"SelectorModeが上書きされるはずです")
		// Commentフィールドは省略されているため、デフォルト値が維持される
		// mergeString は空文字列の場合は上書きしない
		assert.Equal(t, "デフォルトコメント", defaultProfile.AI.Mock.Comment,
			"省略されたCommentはデフォルト値が維持されるはずです")
	})
}
