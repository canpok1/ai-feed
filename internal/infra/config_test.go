package infra

import (
	"os"
	"path/filepath"
	"testing"

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

	// Test Save
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

	// Test Load
	loadedConfig, err := repo.Load()
	assert.NoError(t, err)
	if diff := deep.Equal(configToSave, loadedConfig); diff != nil {
		t.Errorf("Loaded config is not equal to saved config: %v", diff)
	}

	// Test Save when file already exists
	err = repo.Save(configToSave)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config file already exists")

	// Test Load with non-existent file
	nonExistentFilePath := filepath.Join(tmpDir, "non_existent.yaml")
	nonExistentRepo := NewYamlConfigRepository(nonExistentFilePath)
	_, err = nonExistentRepo.Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestYamlConfigRepository_Save_InvalidPath(t *testing.T) {
	invalidPath := "/nonexistent_dir/test_config.yaml" // This path should not exist and cause an error
	repo := NewYamlConfigRepository(invalidPath)

	configToSave := MakeDefaultConfig()
	err := repo.Save(configToSave)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create config file")
}

func TestYamlConfigRepository_Load_InvalidYaml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid_config.yaml")

	// Write invalid YAML content to the file
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualYaml, err := yaml.Marshal(tt.input)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				// Unmarshal the actual YAML back into an OutputConfig to compare
				// since YAML marshaling order is not guaranteed.
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

			entity, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, entity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, tt.config.Type, entity.Type)

				if tt.config.APIKey != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIKey, entity.APIKey)
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, entity.APIKey)
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

			entity, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, entity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, tt.config.Channel, entity.Channel)
				assert.Equal(t, tt.config.MessageTemplate, entity.MessageTemplate)

				if tt.config.APIToken != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIToken, entity.APIToken)
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, entity.APIToken)
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

			entity, err := tt.config.ToEntity()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, entity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, tt.config.APIURL, entity.APIURL)

				if tt.config.APIToken != "" {
					// 直接指定が優先される
					assert.Equal(t, tt.config.APIToken, entity.APIToken)
				} else {
					// 環境変数から取得
					assert.Equal(t, tt.envValue, entity.APIToken)
				}
			}
		})
	}
}
