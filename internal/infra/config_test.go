package infra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewYamlConfigRepository(t *testing.T) {
	repo := NewYamlConfigRepository("test_path.yaml")
	assert.NotNil(t, repo)
	assert.IsType(t, &YamlConfigRepository{}, repo)
}

func TestYamlConfigRepository_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_config.yaml")

	repo := NewYamlConfigRepository(filePath)

	// 保存テスト
	configToSave := &Config{
		DefaultProfile: &Profile{
			AI: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "gemini-test",
					APIKey: "test_api_key",
				},
			},
			Prompt: &PromptConfig{
				SystemPrompt:          "test system prompt",
				CommentPromptTemplate: "test comment prompt template",
				SelectorPrompt:        "test selector prompt",
			},
			Output: &OutputConfig{
				SlackAPI: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "test_channel",
				},
				Misskey: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "http://test.misskey.com",
				},
			},
		},
	}

	err := repo.Save(configToSave)
	assert.NoError(t, err)
	assert.FileExists(t, filePath)

	// 読み込みテスト
	loadedConfig, err := repo.Load()
	assert.NoError(t, err)
	if diff := deep.Equal(configToSave, loadedConfig); diff != nil {
		t.Errorf("Loaded config is not equal to saved config: %v", diff)
	}

	// ファイルが既に存在する場合の保存テスト
	err = repo.Save(configToSave)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file already exists")

	// 存在しないファイルの読み込みテスト
	nonExistentFilePath := filepath.Join(tmpDir, "non_existent.yaml")
	nonExistentRepo := NewYamlConfigRepository(nonExistentFilePath)
	_, err = nonExistentRepo.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestYamlConfigRepository_Save_InvalidPath(t *testing.T) {
	invalidPath := "/nonexistent_dir/test_config.yaml" // このパスは存在せずエラーになるべき
	repo := NewYamlConfigRepository(invalidPath)

	configToSave := &Config{
		DefaultProfile: &Profile{
			AI: &AIConfig{
				Gemini: &GeminiConfig{
					Type:   "gemini-test",
					APIKey: "test-api-key",
				},
			},
		},
	}
	err := repo.Save(configToSave)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config file")
}

func TestYamlConfigRepository_Load_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid_config.yaml")

	// 不正なYAML内容をファイルに書き込む
	err := os.WriteFile(filePath, []byte("invalid: - yaml"), 0644)
	assert.NoError(t, err)

	repo := NewYamlConfigRepository(filePath)
	_, err = repo.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal YAML")
}

func TestOutputConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name        string
		yamlInput   string
		expected    OutputConfig
		expectedErr string
	}{
		{
			name: "misskey type",
			yamlInput: `
misskey:
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expected: OutputConfig{
				Misskey: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
		{
			name: "slack-api type",
			yamlInput: `
slack_api:
  api_token: test_slack_token
  channel: "#general"
`,
			expected: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "slack-api with enabled: true",
			yamlInput: `
slack_api:
  enabled: true
  api_token: test_slack_token
  channel: "#general"
`,
			expected: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "slack-api with enabled: false",
			yamlInput: `
slack_api:
  enabled: false
  api_token: test_slack_token
  channel: "#general"
`,
			expected: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "misskey with enabled: true",
			yamlInput: `
misskey:
  enabled: true
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expected: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
		{
			name: "misskey with enabled: false",
			yamlInput: `
misskey:
  enabled: false
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expected: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual OutputConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &actual)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.Misskey, actual.Misskey)
				assert.Equal(t, tt.expected.SlackAPI, actual.SlackAPI)
			}
		})
	}
}

func TestOutputConfig_MarshalYAML(t *testing.T) {
	tests := []struct {
		name         string
		input        OutputConfig
		expectedYaml string
		expectedErr  string
	}{
		{
			name: "misskey type",
			input: OutputConfig{
				Misskey: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedYaml: `misskey:
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedErr: "",
		},
		{
			name: "slack-api type",
			input: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedYaml: `slack_api:
  api_token: test_slack_token
  channel: "#general"
`,
			expectedErr: "",
		},
		{
			name: "slack-api with enabled: true",
			input: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "slack-api with enabled: false",
			input: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "misskey with enabled: true",
			input: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
		{
			name: "misskey with enabled: false",
			input: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualYaml, err := yaml.Marshal(tt.input)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				// YAMLのマーシャル順序は保証されないため、
				// 実際のYAMLをOutputConfigにアンマーシャルして比較する
				var actualOutput OutputConfig
				err = yaml.Unmarshal(actualYaml, &actualOutput)
				assert.NoError(t, err)
				assert.Equal(t, tt.input.Misskey, actualOutput.Misskey)
				assert.Equal(t, tt.input.SlackAPI, actualOutput.SlackAPI)
			}
		})
	}
}

// 環境変数からのAPIキー/トークン取得機能のテスト

func TestGeminiConfig_ToEntity_WithEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name          string
		config        GeminiConfig
		envVar        string
		envValue      string
		expectedError string
	}{
		{
			name: "APIKey直接指定優先",
			config: GeminiConfig{
				Type:      "gemini-test",
				APIKey:    "direct-key",
				APIKeyEnv: "TEST_GEMINI_KEY",
			},
			envVar:        "TEST_GEMINI_KEY",
			envValue:      "env-key",
			expectedError: "",
		},
		{
			name: "環境変数から取得成功",
			config: GeminiConfig{
				Type:      "gemini-test",
				APIKey:    "",
				APIKeyEnv: "TEST_GEMINI_KEY",
			},
			envVar:        "TEST_GEMINI_KEY",
			envValue:      "env-key-value",
			expectedError: "",
		},
		{
			name: "環境変数が存在しない",
			config: GeminiConfig{
				Type:      "gemini-test",
				APIKey:    "",
				APIKeyEnv: "NON_EXISTENT_KEY",
			},
			envVar:        "",
			envValue:      "",
			expectedError: "環境変数 'NON_EXISTENT_KEY' が設定されていません",
		},
		{
			name: "環境変数が空文字列",
			config: GeminiConfig{
				Type:      "gemini-test",
				APIKey:    "",
				APIKeyEnv: "EMPTY_GEMINI_KEY",
			},
			envVar:        "EMPTY_GEMINI_KEY",
			envValue:      "",
			expectedError: "環境変数 'EMPTY_GEMINI_KEY' が設定されていません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
			}

			got, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.config.Type, got.Type)

				if tt.config.APIKey != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIKey, got.APIKey.Value())
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, got.APIKey.Value())
				}
			}
		})
	}
}

func TestSlackAPIConfig_ToEntity_WithEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name          string
		config        SlackAPIConfig
		envVar        string
		envValue      string
		expectedError string
	}{
		{
			name: "APIToken直接指定優先",
			config: SlackAPIConfig{
				APIToken:    "direct-token",
				APITokenEnv: "TEST_SLACK_TOKEN",
				Channel:     "#test",
			},
			envVar:        "TEST_SLACK_TOKEN",
			envValue:      "env-token",
			expectedError: "",
		},
		{
			name: "環境変数から取得成功",
			config: SlackAPIConfig{
				APIToken:    "",
				APITokenEnv: "TEST_SLACK_TOKEN",
				Channel:     "#test",
			},
			envVar:        "TEST_SLACK_TOKEN",
			envValue:      "env-token-value",
			expectedError: "",
		},
		{
			name: "環境変数が存在しない",
			config: SlackAPIConfig{
				APIToken:    "",
				APITokenEnv: "NON_EXISTENT_SLACK",
				Channel:     "#test",
			},
			envVar:        "",
			envValue:      "",
			expectedError: "環境変数 'NON_EXISTENT_SLACK' が設定されていません",
		},
		{
			name: "環境変数が空文字列",
			config: SlackAPIConfig{
				APIToken:    "",
				APITokenEnv: "EMPTY_SLACK_TOKEN",
				Channel:     "#test",
			},
			envVar:        "EMPTY_SLACK_TOKEN",
			envValue:      "",
			expectedError: "環境変数 'EMPTY_SLACK_TOKEN' が設定されていません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
			}

			got, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.config.Channel, got.Channel)
				assert.Equal(t, tt.config.MessageTemplate, got.MessageTemplate)

				// Enabledフィールドの後方互換性チェック（省略時=true）
				assert.True(t, got.Enabled, "Enabled should default to true for backward compatibility")

				if tt.config.APIToken != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIToken, got.APIToken.Value())
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, got.APIToken.Value())
				}
			}
		})
	}
}

func TestMisskeyConfig_ToEntity_WithEnvironmentVariable(t *testing.T) {
	tests := []struct {
		name          string
		config        MisskeyConfig
		envVar        string
		envValue      string
		expectedError string
	}{
		{
			name: "APIToken直接指定優先",
			config: MisskeyConfig{
				APIToken:    "direct-token",
				APITokenEnv: "TEST_MISSKEY_TOKEN",
				APIURL:      "https://test.misskey.com",
			},
			envVar:        "TEST_MISSKEY_TOKEN",
			envValue:      "env-token",
			expectedError: "",
		},
		{
			name: "環境変数から取得成功",
			config: MisskeyConfig{
				APIToken:    "",
				APITokenEnv: "TEST_MISSKEY_TOKEN",
				APIURL:      "https://test.misskey.com",
			},
			envVar:        "TEST_MISSKEY_TOKEN",
			envValue:      "env-token-value",
			expectedError: "",
		},
		{
			name: "環境変数が存在しない",
			config: MisskeyConfig{
				APIToken:    "",
				APITokenEnv: "NON_EXISTENT_MISSKEY",
				APIURL:      "https://test.misskey.com",
			},
			envVar:        "",
			envValue:      "",
			expectedError: "環境変数 'NON_EXISTENT_MISSKEY' が設定されていません",
		},
		{
			name: "環境変数が空文字列",
			config: MisskeyConfig{
				APIToken:    "",
				APITokenEnv: "EMPTY_MISSKEY_TOKEN",
				APIURL:      "https://test.misskey.com",
			},
			envVar:        "EMPTY_MISSKEY_TOKEN",
			envValue:      "",
			expectedError: "環境変数 'EMPTY_MISSKEY_TOKEN' が設定されていません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
			}

			got, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.config.APIURL, got.APIURL)

				// Enabledフィールドの後方互換性チェック（省略時=true）
				assert.True(t, got.Enabled, "Enabled should default to true for backward compatibility")

				if tt.config.APIToken != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIToken, got.APIToken.Value())
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, got.APIToken.Value())
				}
			}
		})
	}
}

func TestSlackAPIConfig_ToEntity_WithEnabledFlag(t *testing.T) {
	tests := []struct {
		name          string
		config        SlackAPIConfig
		expectEnabled bool
	}{
		{
			name: "enabled省略時（後方互換性）",
			config: SlackAPIConfig{
				APIToken: "test_token",
				Channel:  "#test",
			},
			expectEnabled: true,
		},
		{
			name: "enabled: true",
			config: SlackAPIConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: "test_token",
				Channel:  "#test",
			},
			expectEnabled: true,
		},
		{
			name: "enabled: false",
			config: SlackAPIConfig{
				Enabled:  testutil.BoolPtr(false),
				APIToken: "test_token",
				Channel:  "#test",
			},
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ToEntity()

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectEnabled, got.Enabled)
		})
	}
}

func TestMisskeyConfig_ToEntity_WithEnabledFlag(t *testing.T) {
	tests := []struct {
		name          string
		config        MisskeyConfig
		expectEnabled bool
	}{
		{
			name: "enabled省略時（後方互換性）",
			config: MisskeyConfig{
				APIToken: "test_token",
				APIURL:   "https://test.misskey.io",
			},
			expectEnabled: true,
		},
		{
			name: "enabled: true",
			config: MisskeyConfig{
				Enabled:  testutil.BoolPtr(true),
				APIToken: "test_token",
				APIURL:   "https://test.misskey.io",
			},
			expectEnabled: true,
		},
		{
			name: "enabled: false",
			config: MisskeyConfig{
				Enabled:  testutil.BoolPtr(false),
				APIToken: "test_token",
				APIURL:   "https://test.misskey.io",
			},
			expectEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ToEntity()

			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.expectEnabled, got.Enabled)
		})
	}
}

// 組み合わせパターンテスト

func TestOutputConfig_EnabledCombinationPatterns(t *testing.T) {
	tests := []struct {
		name                   string
		yamlInput              string
		expectedSlackEnabled   *bool
		expectedMisskeyEnabled *bool
		expectedErr            string
	}{
		{
			name: "両方省略（後方互換性）",
			yamlInput: `
slack_api:
  api_token: test_slack_token
  channel: "#general"
misskey:
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedSlackEnabled:   nil,
			expectedMisskeyEnabled: nil,
			expectedErr:            "",
		},
		{
			name: "両方有効（explicit true）",
			yamlInput: `
slack_api:
  enabled: true
  api_token: test_slack_token
  channel: "#general"
misskey:
  enabled: true
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedSlackEnabled:   testutil.BoolPtr(true),
			expectedMisskeyEnabled: testutil.BoolPtr(true),
			expectedErr:            "",
		},
		{
			name: "SlackAPI有効、Misskey無効",
			yamlInput: `
slack_api:
  enabled: true
  api_token: test_slack_token
  channel: "#general"
misskey:
  enabled: false
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedSlackEnabled:   testutil.BoolPtr(true),
			expectedMisskeyEnabled: testutil.BoolPtr(false),
			expectedErr:            "",
		},
		{
			name: "SlackAPI無効、Misskey有効",
			yamlInput: `
slack_api:
  enabled: false
  api_token: test_slack_token
  channel: "#general"
misskey:
  enabled: true
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedSlackEnabled:   testutil.BoolPtr(false),
			expectedMisskeyEnabled: testutil.BoolPtr(true),
			expectedErr:            "",
		},
		{
			name: "両方無効",
			yamlInput: `
slack_api:
  enabled: false
  api_token: test_slack_token
  channel: "#general"
misskey:
  enabled: false
  api_token: test_misskey_token
  api_url: https://misskey.example.com
`,
			expectedSlackEnabled:   testutil.BoolPtr(false),
			expectedMisskeyEnabled: testutil.BoolPtr(false),
			expectedErr:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual OutputConfig
			err := yaml.Unmarshal([]byte(tt.yamlInput), &actual)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)

				// SlackAPIのEnabledフィールドチェック
				if tt.expectedSlackEnabled == nil {
					assert.Nil(t, actual.SlackAPI.Enabled)
				} else {
					assert.NotNil(t, actual.SlackAPI.Enabled)
					assert.Equal(t, *tt.expectedSlackEnabled, *actual.SlackAPI.Enabled)
				}

				// MisskeyのEnabledフィールドチェック
				if tt.expectedMisskeyEnabled == nil {
					assert.Nil(t, actual.Misskey.Enabled)
				} else {
					assert.NotNil(t, actual.Misskey.Enabled)
					assert.Equal(t, *tt.expectedMisskeyEnabled, *actual.Misskey.Enabled)
				}
			}
		})
	}
}

func TestOutputConfig_ToEntity_EnabledCombinations(t *testing.T) {
	tests := []struct {
		name                   string
		config                 OutputConfig
		expectedSlackEnabled   bool
		expectedMisskeyEnabled bool
		expectedErr            string
	}{
		{
			name: "両方省略（後方互換性）",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedSlackEnabled:   true,
			expectedMisskeyEnabled: true,
			expectedErr:            "",
		},
		{
			name: "両方有効",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedSlackEnabled:   true,
			expectedMisskeyEnabled: true,
			expectedErr:            "",
		},
		{
			name: "SlackAPI有効、Misskey無効",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedSlackEnabled:   true,
			expectedMisskeyEnabled: false,
			expectedErr:            "",
		},
		{
			name: "SlackAPI無効、Misskey有効",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(true),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedSlackEnabled:   false,
			expectedMisskeyEnabled: true,
			expectedErr:            "",
		},
		{
			name: "両方無効",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
				Misskey: &MisskeyConfig{
					Enabled:  testutil.BoolPtr(false),
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedSlackEnabled:   false,
			expectedMisskeyEnabled: false,
			expectedErr:            "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ToEntity()

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// SlackAPIのEnabledフィールドチェック
				if got.SlackAPI != nil {
					assert.Equal(t, tt.expectedSlackEnabled, got.SlackAPI.Enabled)
				}

				// MisskeyのEnabledフィールドチェック
				if got.Misskey != nil {
					assert.Equal(t, tt.expectedMisskeyEnabled, got.Misskey.Enabled)
				}
			}
		})
	}
}

func TestOutputConfig_EnabledFalseWithValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		config      OutputConfig
		expectedErr string
	}{
		{
			name: "SlackAPI無効の場合はAPIトークンのバリデーションをスキップ",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:     testutil.BoolPtr(false),
					APIToken:    "", // 空でもエラーにならない
					APITokenEnv: "NON_EXISTENT_TOKEN",
					Channel:     "#general",
				},
			},
			expectedErr: "", // エラーなし
		},
		{
			name: "Misskey無効の場合はAPIトークンのバリデーションをスキップ",
			config: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:     testutil.BoolPtr(false),
					APIToken:    "", // 空でもエラーにならない
					APITokenEnv: "NON_EXISTENT_TOKEN",
					APIURL:      "https://misskey.example.com",
				},
			},
			expectedErr: "", // エラーなし
		},
		{
			name: "SlackAPI有効の場合はAPIトークンが必要",
			config: OutputConfig{
				SlackAPI: &SlackAPIConfig{
					Enabled:     testutil.BoolPtr(true),
					APIToken:    "", // 空
					APITokenEnv: "NON_EXISTENT_TOKEN",
					Channel:     "#general",
				},
			},
			expectedErr: "環境変数",
		},
		{
			name: "Misskey有効の場合はAPIトークンが必要",
			config: OutputConfig{
				Misskey: &MisskeyConfig{
					Enabled:     testutil.BoolPtr(true),
					APIToken:    "", // 空
					APITokenEnv: "NON_EXISTENT_TOKEN",
					APIURL:      "https://misskey.example.com",
				},
			},
			expectedErr: "環境変数",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.ToEntity()

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// キャッシュ設定テスト

func TestCacheConfig_ToEntity(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		config      CacheConfig
		expected    *entity.CacheConfig
		expectedErr string
	}{
		{
			name: "正常な設定値での変換",
			config: CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/tmp/cache.jsonl",
				MaxEntries:    1000,
				RetentionDays: 30,
			},
			expected: &entity.CacheConfig{
				Enabled:       true,
				FilePath:      "/tmp/cache.jsonl",
				MaxEntries:    1000,
				RetentionDays: 30,
			},
			expectedErr: "",
		},
		{
			name: "デフォルト値の適用",
			config: CacheConfig{
				Enabled:       nil,
				FilePath:      "",
				MaxEntries:    0,
				RetentionDays: 0,
			},
			expected: &entity.CacheConfig{
				Enabled:       false,
				FilePath:      filepath.Join(homeDir, ".ai-feed", "recommend_history.jsonl"),
				MaxEntries:    1000,
				RetentionDays: 30,
			},
			expectedErr: "",
		},
		{
			name: "チルダ展開テスト",
			config: CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "~/.ai-feed/custom-cache.jsonl",
				MaxEntries:    500,
				RetentionDays: 15,
			},
			expected: &entity.CacheConfig{
				Enabled:       true,
				FilePath:      filepath.Join(homeDir, ".ai-feed", "custom-cache.jsonl"),
				MaxEntries:    500,
				RetentionDays: 15,
			},
			expectedErr: "",
		},
		{
			name: "相対パス展開テスト",
			config: CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "./cache/data.jsonl",
				MaxEntries:    2000,
				RetentionDays: 60,
			},
			expected: &entity.CacheConfig{
				Enabled:       false,
				FilePath:      filepath.Join(".", "cache", "data.jsonl"),
				MaxEntries:    2000,
				RetentionDays: 60,
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.config.ToEntity()

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Enabled, result.Enabled)
				// パスは絶対パスに展開されるため、期待値も絶対パスに変換して比較
				expectedPath, _ := filepath.Abs(tt.expected.FilePath)
				actualPath, _ := filepath.Abs(result.FilePath)
				assert.Equal(t, expectedPath, actualPath)
				assert.Equal(t, tt.expected.MaxEntries, result.MaxEntries)
				assert.Equal(t, tt.expected.RetentionDays, result.RetentionDays)
			}
		})
	}
}

// TestProfile_ToEntity_WithCache は削除 - ProfileはキャッシュfieldBはなくなったため

func TestCacheConfig_YamlMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		expected  CacheConfig
	}{
		{
			name: "完全な設定",
			yamlInput: `
cache:
  enabled: true
  file_path: /tmp/cache.jsonl
  max_entries: 1000
  retention_days: 30
`,
			expected: CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "/tmp/cache.jsonl",
				MaxEntries:    1000,
				RetentionDays: 30,
			},
		},
		{
			name: "enabled無効",
			yamlInput: `
cache:
  enabled: false
  file_path: ~/.ai-feed/disabled-cache.jsonl
  max_entries: 0
  retention_days: 0
`,
			expected: CacheConfig{
				Enabled:       testutil.BoolPtr(false),
				FilePath:      "~/.ai-feed/disabled-cache.jsonl",
				MaxEntries:    0,
				RetentionDays: 0,
			},
		},
		{
			name: "最小設定",
			yamlInput: `
cache:
  enabled: true
`,
			expected: CacheConfig{
				Enabled:       testutil.BoolPtr(true),
				FilePath:      "",
				MaxEntries:    0,
				RetentionDays: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type TestStruct struct {
				Cache CacheConfig `yaml:"cache"`
			}

			var actual TestStruct
			err := yaml.Unmarshal([]byte(tt.yamlInput), &actual)
			assert.NoError(t, err)

			if tt.expected.Enabled == nil {
				assert.Nil(t, actual.Cache.Enabled)
			} else {
				assert.NotNil(t, actual.Cache.Enabled)
				assert.Equal(t, *tt.expected.Enabled, *actual.Cache.Enabled)
			}
			assert.Equal(t, tt.expected.FilePath, actual.Cache.FilePath)
			assert.Equal(t, tt.expected.MaxEntries, actual.Cache.MaxEntries)
			assert.Equal(t, tt.expected.RetentionDays, actual.Cache.RetentionDays)

			// Marshal テスト
			marshaledYaml, err := yaml.Marshal(actual)
			assert.NoError(t, err)

			var remarshaledActual TestStruct
			err = yaml.Unmarshal(marshaledYaml, &remarshaledActual)
			assert.NoError(t, err)
			assert.Equal(t, actual.Cache.Enabled, remarshaledActual.Cache.Enabled)
			assert.Equal(t, actual.Cache.FilePath, remarshaledActual.Cache.FilePath)
			assert.Equal(t, actual.Cache.MaxEntries, remarshaledActual.Cache.MaxEntries)
			assert.Equal(t, actual.Cache.RetentionDays, remarshaledActual.Cache.RetentionDays)
		})
	}
}
