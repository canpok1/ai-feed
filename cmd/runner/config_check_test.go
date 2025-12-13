package runner

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigCheckRunner(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
	}{
		{
			name:       "正常系: 新しいインスタンスを作成できる",
			configPath: "config.yml",
		},
		{
			name:       "正常系: 空のconfigPathでも作成できる",
			configPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			runner := NewConfigCheckRunner(tt.configPath, stdout, stderr)

			require.NotNil(t, runner)
			assert.Equal(t, tt.configPath, runner.configPath)
			assert.Equal(t, stdout, runner.stdout)
			assert.Equal(t, stderr, runner.stderr)
		})
	}
}
