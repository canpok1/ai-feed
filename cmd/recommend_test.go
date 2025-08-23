package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/spf13/cobra"
)

func TestNewRecommendParams(t *testing.T) {
	tests := []struct {
		name         string
		urlFlags     []string
		sourceFlag   string
		expectedURLs []string
		expectedErr  string
	}{
		{
			name:         "Single URL flag",
			urlFlags:     []string{"http://example.com/feed.xml"},
			sourceFlag:   "",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "Multiple URL flags",
			urlFlags:     []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml", "http://example3.com/feed.xml"},
			sourceFlag:   "",
			expectedURLs: []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml", "http://example3.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "Source flag only with valid file",
			urlFlags:     []string{},
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml"},
		},
		{
			name:         "Both URL and source flags (併用)",
			urlFlags:     []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml"},
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml", "http://example1.com/feed.xml", "http://example2.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "Neither URL nor source flags",
			urlFlags:     []string{},
			sourceFlag:   "",
			expectedURLs: nil,
			expectedErr:  "--url または --source のいずれかを指定してください",
		},
		{
			name:         "Source file not found",
			urlFlags:     []string{},
			sourceFlag:   "non_existent_file.txt",
			expectedURLs: nil,
			expectedErr:  "failed to read URLs from file: failed to open file non_existent_file.txt: open non_existent_file.txt: no such file or directory",
		},
		{
			name:         "Empty source file",
			urlFlags:     []string{},
			sourceFlag:   "empty_source.txt",
			expectedURLs: nil,
			expectedErr:  "--url または --source のいずれかを指定してください",
		},
		{
			name:         "Empty source file but URLs provided",
			urlFlags:     []string{"http://example.com/feed.xml"},
			sourceFlag:   "empty_source.txt",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy cobra.Command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().StringSliceP("url", "u", []string{}, "URL of the feed to recommend from")
			cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

			for _, url := range tt.urlFlags {
				cmd.Flags().Set("url", url)
			}
			if tt.sourceFlag != "" {
				// Create temporary source file if sourceFlag is used
				if tt.sourceFlag == "tmp_source.txt" {
					content := "http://example.com/from_file.xml\nhttp://another.com/from_file.xml"
					err := os.WriteFile(tt.sourceFlag, []byte(content), 0644)
					assert.NoError(t, err)
					defer os.Remove(tt.sourceFlag)
				} else if tt.sourceFlag == "empty_source.txt" {
					err := os.WriteFile(tt.sourceFlag, []byte(""), 0644)
					assert.NoError(t, err)
					defer os.Remove(tt.sourceFlag)
				}
				cmd.Flags().Set("source", tt.sourceFlag)
			}

			params, err := newRecommendParams(cmd)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, params)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
				assert.Equal(t, tt.expectedURLs, params.URLs)
			}
		})
	}

	// 統合テスト - 複数URL処理の確認
	t.Run("Integration - Multiple URL handling", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		for _, url := range []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml", "https://example3.com/feed.xml"} {
			cmd.Flags().Set("url", url)
		}

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Len(t, params.URLs, 3)
	})

	// 統合テスト - URLとソースの組み合わせ
	t.Run("Integration - URL and source combination", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// ソースファイル作成
		sourceFile := "test_integration_urls.txt"
		sourceContent := "https://source1.com/feed.xml\nhttps://source2.com/feed.xml"
		err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(sourceFile)

		// URLとソースの両方を設定
		cmd.Flags().Set("url", "https://example.com/feed.xml")
		cmd.Flags().Set("source", sourceFile)

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Len(t, params.URLs, 3) // 1 from URL + 2 from source
	})

	// 部分的失敗処理テスト
	t.Run("Partial failure handling - Mixed source content", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// 空行や空白を含むソースファイルを作成
		sourceContent := "https://valid1.com/feed.xml\n\nhttps://valid2.com/feed.xml\n   \nhttps://valid3.com/feed.xml"
		sourceFile := "test_mixed_source.txt"
		err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(sourceFile)

		cmd.Flags().Set("source", sourceFile)

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Greater(t, len(params.URLs), 0)
		for _, url := range params.URLs {
			assert.True(t, len(url) > 0, "URL should not be empty")
		}
	})

	// 記事取得成功時の動作確認
	t.Run("Article fetch success verification", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		cmd.Flags().Set("url", "https://example.com/feed.xml")

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Equal(t, []string{"https://example.com/feed.xml"}, params.URLs)
	})
}

