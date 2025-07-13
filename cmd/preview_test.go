package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTempFile はテスト用のテンポラリファイルを作成し、そのパスを返します。
func createTempFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "test_urls_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpfile.Name()
}

func TestReadURLsFromFile(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedURLs  []string
		expectError   bool
		expectWarning bool
		fileName      string // 存在しないファイルテスト用
	}{
		{
			name:          "有効なURLリスト",
			content:       "http://example.com/1\nhttps://example.org/2\n",
			expectedURLs:  []string{"http://example.com/1", "https://example.org/2"},
			expectError:   false,
			expectWarning: false,
		},
		{
			name:          "空行を含むURLリスト",
			content:       "http://example.com/1\n\nhttps://example.org/2\n\n",
			expectedURLs:  []string{"http://example.com/1", "https://example.org/2"},
			expectError:   false,
			expectWarning: false,
		},
		{
			name:          "不正なURLを含むURLリスト",
			content:       "http://example.com/1\nnot-a-url\nhttps://example.org/2\n",
			expectedURLs:  []string{"http://example.com/1", "https://example.org/2"},
			expectError:   false,
			expectWarning: true,
		},
		{
			name:          "存在しないファイル",
			content:       "",
			expectedURLs:  nil,
			expectError:   true,
			expectWarning: false,
			fileName:      "non_existent_file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.fileName != "" {
				filePath = tt.fileName
			} else {
				filePath = createTempFile(t, tt.content)
				defer os.Remove(filePath)
			}

			// 標準エラー出力をキャプチャ
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			cmd := &cobra.Command{} // モックのcobra.Command
			urls, err := readURLsFromFile(filePath, cmd)

			w.Close()
			os.Stderr = oldStderr // 標準エラー出力を元に戻す
			var buf bytes.Buffer
			buf.ReadFrom(r)
			stderrOutput := buf.String()

			if (err != nil) != tt.expectError {
				t.Fatalf("readURLsFromFile() error = %v, expectError %v", err, tt.expectError)
			}

			if tt.expectError {
				if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
					t.Errorf("Expected 'no such file or directory' error, got: %v", err)
				}
				return
			}

			if len(urls) != len(tt.expectedURLs) {
				t.Fatalf("Expected %d URLs, got %d", len(tt.expectedURLs), len(urls))
			}
			for i, u := range urls {
				if u != tt.expectedURLs[i] {
					t.Errorf("Expected URL %s, got %s", tt.expectedURLs[i], u)
				}
			}

			if tt.expectWarning && (!strings.Contains(stderrOutput, "Warning: Invalid URL in") || !strings.Contains(stderrOutput, "not-a-url")) {
				t.Errorf("Expected warning for invalid URL, got: %s", stderrOutput)
			} else if !tt.expectWarning && stderrOutput != "" {
				t.Errorf("Unexpected warning output: %s", stderrOutput)
			}
		})
	}
}

// TestPreviewCommandSourceAndURLConflict は --source と --url オプションの同時使用をテストします。
func TestPreviewCommandSourceAndURLConflict(t *testing.T) {
	// Use the actual previewCmd, which has its flags defined in its init() function.
	cmd := previewCmd
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(b)

	// Reset flags to avoid state leakage from other tests.
	cmd.Flags().Set("url", "")
	cmd.Flags().Set("source", "")

	// Simulate user providing flags via arguments.
	args := []string{"--url", "http://example.com", "--source", "list.txt"}
	// Manually parse flags to correctly set the "Changed" status.
	if err := cmd.ParseFlags(args); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	err := cmd.RunE(cmd, args)
	if err == nil {
		t.Fatal("Expected an error when --source and --url are used together, but got none.")
	}
	if !strings.Contains(err.Error(), "cannot use --source and --url options together") {
		t.Errorf("Expected conflict error, got: %v", err)
	}
}

// TestPreviewCommandDuplicateURLs はURLの重複排除をテストします。
func TestPreviewCommandDuplicateURLs(t *testing.T) {
	tests := []struct {
		name         string
		inputURLs    []string
		expectedURLs []string
	}{
		{
			name:         "重複なし",
			inputURLs:    []string{"http://example.com/1", "http://example.com/2"},
			expectedURLs: []string{"http://example.com/1", "http://example.com/2"},
		},
		{
			name:         "重複あり",
			inputURLs:    []string{"http://example.com/1", "http://example.com/2", "http://example.com/1"},
			expectedURLs: []string{"http://example.com/1", "http://example.com/2"},
		},
		{
			name:         "空のリスト",
			inputURLs:    nil,
			expectedURLs: nil,
		},
		{
			name:         "すべて重複",
			inputURLs:    []string{"http://example.com/1", "http://example.com/1", "http://example.com/1"},
			expectedURLs: []string{"http://example.com/1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualURLs := deduplicateURLs(tt.inputURLs)

			if len(actualURLs) != len(tt.expectedURLs) {
				t.Fatalf("Expected %d unique URLs, got %d", len(tt.expectedURLs), len(actualURLs))
			}
			for i, u := range actualURLs {
				if u != tt.expectedURLs[i] {
					t.Errorf("Expected unique URL %s, got %s", tt.expectedURLs[i], u)
				}
			}
		})
	}
}
