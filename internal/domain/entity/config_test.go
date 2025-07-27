package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

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
			expectedYaml: `type: misskey
api_token: test_misskey_token
api_url: https://misskey.example.com
`,
			expectedErr: "",
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
			expectedYaml: `type: slack-api
api_token: test_slack_token
channel: "#general"
`,
			expectedErr: "",
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
