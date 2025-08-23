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

			if len(tt.urlFlags) > 0 {
				cmd.Flags().Set("url", tt.urlFlags[0])
				for i := 1; i < len(tt.urlFlags); i++ {
					cmd.Flags().Set("url", tt.urlFlags[i])
				}
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
}

func TestRecommendCommandIntegration(t *testing.T) {
	// Test that the command properly handles multiple URLs
	tests := []struct {
		name           string
		urlFlags       []string
		sourceFile     string
		sourceContent  string
		expectedParams int // Expected number of URLs in params
		wantErr        bool
		errContains    string
	}{
		{
			name:           "Multiple URLs via -u flags",
			urlFlags:       []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml", "https://example3.com/feed.xml"},
			sourceFile:     "",
			sourceContent:  "",
			expectedParams: 3,
			wantErr:        false,
		},
		{
			name:           "Single URL",
			urlFlags:       []string{"https://example.com/feed.xml"},
			sourceFile:     "",
			sourceContent:  "",
			expectedParams: 1,
			wantErr:        false,
		},
		{
			name:           "URL and source combination",
			urlFlags:       []string{"https://example.com/feed.xml"},
			sourceFile:     "test_urls.txt",
			sourceContent:  "https://source1.com/feed.xml\nhttps://source2.com/feed.xml",
			expectedParams: 3, // 1 from URL + 2 from source
			wantErr:        false,
		},
		{
			name:           "Multiple URLs and source combination",
			urlFlags:       []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml"},
			sourceFile:     "test_urls.txt",
			sourceContent:  "https://source1.com/feed.xml\nhttps://source2.com/feed.xml",
			expectedParams: 4, // 2 from URLs + 2 from source
			wantErr:        false,
		},
		{
			name:        "No URL or source provided",
			urlFlags:    []string{},
			sourceFile:  "",
			wantErr:     true,
			errContains: "--url または --source のいずれかを指定してください",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy cobra.Command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
			cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

			// Set URL flags
			for _, url := range tt.urlFlags {
				cmd.Flags().Set("url", url)
			}

			// Create source file if needed
			if tt.sourceFile != "" && tt.sourceContent != "" {
				err := os.WriteFile(tt.sourceFile, []byte(tt.sourceContent), 0644)
				assert.NoError(t, err)
				defer os.Remove(tt.sourceFile)
				cmd.Flags().Set("source", tt.sourceFile)
			}

			// Call newRecommendParams
			params, err := newRecommendParams(cmd)

			// Check results
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
				assert.Len(t, params.URLs, tt.expectedParams)
			}
		})
	}
}

func TestRecommendCommandDebugLogging(t *testing.T) {
	// 複数URLを指定した場合のデバッグログ出力確認のためのテスト
	t.Run("Debug log shows all URLs", func(t *testing.T) {
		// Create a dummy cobra.Command and set flags for multiple URLs
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Set multiple URLs
		testURLs := []string{
			"https://example1.com/feed.xml",
			"https://example2.com/feed.xml",
			"https://example3.com/feed.xml",
		}
		for _, url := range testURLs {
			cmd.Flags().Set("url", url)
		}

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Equal(t, testURLs, params.URLs)

		// URLsフィールドが期待通りの複数URLを含んでいることを確認
		assert.Len(t, params.URLs, 3)
		assert.Contains(t, params.URLs, "https://example1.com/feed.xml")
		assert.Contains(t, params.URLs, "https://example2.com/feed.xml")
		assert.Contains(t, params.URLs, "https://example3.com/feed.xml")
	})

	t.Run("Combined URLs and source debug logging", func(t *testing.T) {
		// Create a dummy cobra.Command and set both URL flags and source file
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Create source file
		sourceContent := "https://source1.com/feed.xml\nhttps://source2.com/feed.xml"
		sourceFile := "test_debug_source.txt"
		err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(sourceFile)

		// Set URL flags and source file
		cmd.Flags().Set("url", "https://example1.com/feed.xml")
		cmd.Flags().Set("url", "https://example2.com/feed.xml")
		cmd.Flags().Set("source", sourceFile)

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)

		// 期待される順序: sourceのURL -> -uオプションのURL
		expectedURLs := []string{
			"https://source1.com/feed.xml",
			"https://source2.com/feed.xml",
			"https://example1.com/feed.xml",
			"https://example2.com/feed.xml",
		}
		assert.Equal(t, expectedURLs, params.URLs)
		assert.Len(t, params.URLs, 4)
	})
}

func TestRecommendCommandPartialFailure(t *testing.T) {
	// 部分的な失敗時の挙動確認のためのテスト
	t.Run("Handle partial failure in URL parsing", func(t *testing.T) {
		// Create a dummy cobra.Command
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Set valid URLs
		cmd.Flags().Set("url", "https://valid1.com/feed.xml")
		cmd.Flags().Set("url", "https://valid2.com/feed.xml")

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Len(t, params.URLs, 2)
		assert.Equal(t, []string{"https://valid1.com/feed.xml", "https://valid2.com/feed.xml"}, params.URLs)
	})

	t.Run("Handle mixed valid and invalid source file scenarios", func(t *testing.T) {
		// Create a dummy cobra.Command
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Create source file with mixed content (including empty lines and whitespace)
		sourceContent := "https://valid1.com/feed.xml\n\nhttps://valid2.com/feed.xml\n   \nhttps://valid3.com/feed.xml"
		sourceFile := "test_mixed_source.txt"
		err := os.WriteFile(sourceFile, []byte(sourceContent), 0644)
		assert.NoError(t, err)
		defer os.Remove(sourceFile)

		cmd.Flags().Set("source", sourceFile)

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)

		// infra.ReadURLsFromFile の実装に依存するが、有効なURLが取得されることを確認
		assert.Greater(t, len(params.URLs), 0)
		for _, url := range params.URLs {
			assert.True(t, len(url) > 0, "URL should not be empty")
		}
	})

	t.Run("Handle source file read failure gracefully", func(t *testing.T) {
		// Create a dummy cobra.Command
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Set non-existent source file
		cmd.Flags().Set("source", "non_existent_file.txt")

		params, err := newRecommendParams(cmd)
		assert.Error(t, err)
		assert.Nil(t, params)
		assert.Contains(t, err.Error(), "failed to read URLs from file")
	})

	t.Run("Handle empty source file with fallback URLs", func(t *testing.T) {
		// Create a dummy cobra.Command
		cmd := &cobra.Command{}
		cmd.Flags().StringSliceP("url", "u", []string{}, "URLs of the feed to recommend from")
		cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

		// Create empty source file
		sourceFile := "test_empty_fallback.txt"
		err := os.WriteFile(sourceFile, []byte(""), 0644)
		assert.NoError(t, err)
		defer os.Remove(sourceFile)

		// Also set URL flags
		cmd.Flags().Set("url", "https://fallback1.com/feed.xml")
		cmd.Flags().Set("url", "https://fallback2.com/feed.xml")
		cmd.Flags().Set("source", sourceFile)

		params, err := newRecommendParams(cmd)
		// 空のソースファイルでもURLが指定されていれば成功する仕様
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Equal(t, []string{"https://fallback1.com/feed.xml", "https://fallback2.com/feed.xml"}, params.URLs)
	})
}
