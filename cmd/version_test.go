package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestVersionCommand はversionコマンドの動作を確認する
func TestVersionCommand(t *testing.T) {
	testCases := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "default version",
			version: "dev",
			want:    "dev",
		},
		{
			name:    "custom version",
			version: "v1.2.3",
			want:    "v1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalVersion := version
			version = tc.version
			defer func() { version = originalVersion }()

			var buf bytes.Buffer
			cmd := makeVersionCmd()
			cmd.SetOut(&buf)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("versionコマンドの実行に失敗: %v", err)
			}

			got := strings.TrimSpace(buf.String())
			if got != tc.want {
				t.Errorf("期待される出力は %q ですが、実際は %q でした", tc.want, got)
			}
		})
	}
}
