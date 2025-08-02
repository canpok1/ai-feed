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
