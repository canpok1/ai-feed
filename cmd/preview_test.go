package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestPreviewCommandSourceAndURLConflict は --source と --url オプションの同時使用をテストします。
func TestPreviewCommandSourceAndURLConflict(t *testing.T) {
	cmd := makePreviewCmd()

	// 標準出力と標準エラーをキャプチャするためのバッファ
	stdout := bytes.NewBufferString("")
	stderr := bytes.NewBufferString("")
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	// コマンドライン引数を設定
	cmd.SetArgs([]string{"--url", "http://example.com", "--source", "list.txt"})

	// コマンドを実行
	_, err := cmd.ExecuteC()

	// エラーが返されることを確認
	if err == nil {
		t.Fatal("Expected an error when --source and --url are used together, but got none.")
	}

	// エラーメッセージが期待通りであることを確認
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
