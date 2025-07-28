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
		General: GeneralConfig{
			DefaultExecutionProfile: "test-profile",
		},
		Cache:         CacheConfig{},
		AIModels:      map[string]AIModelConfig{},
		SystemPrompts: map[string]string{},
		Prompts:       map[string]PromptConfig{},
		Outputs: map[string]OutputConfig{
			"test-output": {
				Type: "misskey",
				MisskeyConfig: &MisskeyConfig{
					APIToken: "test_token",
					APIURL:   "http://test.misskey.com",
				},
			},
		},
		ExecutionProfiles: map[string]ExecutionProfile{},
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
	assert.Contains(t, err.Error(), "failed to unmarshal config")
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
type: misskey
api_token: test_misskey_token
api_url: https://misskey.example.com
`,
			expected: OutputConfig{
				Type: "misskey",
				MisskeyConfig: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedErr: "",
		},
		{
			name: "slack-api type",
			yamlInput: `
type: slack-api
api_token: test_slack_token
channel: "#general"
`,
			expected: OutputConfig{
				Type: "slack-api",
				SlackAPIConfig: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedErr: "",
		},
		{
			name: "unsupported type",
			yamlInput: `
type: unknown
`,
			expected: OutputConfig{
				Type: "unknown",
			},
			expectedErr: "unsupported output type: unknown",
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
				assert.Equal(t, tt.expected.Type, actual.Type)
				assert.Equal(t, tt.expected.MisskeyConfig, actual.MisskeyConfig)
				assert.Equal(t, tt.expected.SlackAPIConfig, actual.SlackAPIConfig)
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
				Type: "misskey",
				MisskeyConfig: &MisskeyConfig{
					APIToken: "test_misskey_token",
					APIURL:   "https://misskey.example.com",
				},
			},
			expectedYaml: `type: misskey\napi_token: test_misskey_token\napi_url: https://misskey.example.com\n`,
			expectedErr:  "",
		},
		{
			name: "slack-api type",
			input: OutputConfig{
				Type: "slack-api",
				SlackAPIConfig: &SlackAPIConfig{
					APIToken: "test_slack_token",
					Channel:  "#general",
				},
			},
			expectedYaml: `type: slack-api\napi_token: test_slack_token\nchannel: "#general"\n`,
			expectedErr:  "",
		},
		{
			name: "unknown type (should fail)",
			input: OutputConfig{
				Type: "unknown",
			},
			expectedYaml: "",
			expectedErr:  "unsupported output type: unknown",
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
				assert.Equal(t, tt.input.Type, actualOutput.Type)
				assert.Equal(t, tt.input.MisskeyConfig, actualOutput.MisskeyConfig)
				assert.Equal(t, tt.input.SlackAPIConfig, actualOutput.SlackAPIConfig)
			}
		})
	}
}
