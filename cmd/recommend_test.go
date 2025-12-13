package cmd

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// newTestRecommendCmd はテスト用のcobra.Commandを作成するヘルパー関数
func newTestRecommendCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringSliceP("url", "u", []string{}, "推薦対象のフィードURL")
	cmd.Flags().StringP("source", "s", "", "URL一覧が記載されたファイルのパス")
	return cmd
}

func TestNewRecommendParams(t *testing.T) {
	tests := []struct {
		name         string
		urlFlags     []string
		sourceFlag   string
		expectedURLs []string
		expectedErr  string
	}{
		{
			name:         "正常系: 単一URLフラグ",
			urlFlags:     []string{"http://example.com/feed.xml"},
			sourceFlag:   "",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "正常系: 複数URLフラグ",
			urlFlags:     []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml", "http://example3.com/feed.xml"},
			sourceFlag:   "",
			expectedURLs: []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml", "http://example3.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "正常系: ソースフラグのみ（有効なファイル）",
			urlFlags:     []string{},
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml"},
		},
		{
			name:         "正常系: URLとソースフラグの併用",
			urlFlags:     []string{"http://example1.com/feed.xml", "http://example2.com/feed.xml"},
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml", "http://example1.com/feed.xml", "http://example2.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "異常系: URLとソースフラグの両方なし",
			urlFlags:     []string{},
			sourceFlag:   "",
			expectedURLs: nil,
			expectedErr:  "--url または --source のいずれかを指定してください",
		},
		{
			name:         "異常系: ソースファイルが見つからない",
			urlFlags:     []string{},
			sourceFlag:   "non_existent_file.txt",
			expectedURLs: nil,
			expectedErr:  "failed to read URLs from file: failed to open file non_existent_file.txt: open non_existent_file.txt: no such file or directory",
		},
		{
			name:         "異常系: 空のソースファイル",
			urlFlags:     []string{},
			sourceFlag:   "empty_source.txt",
			expectedURLs: nil,
			expectedErr:  "--url または --source のいずれかを指定してください",
		},
		{
			name:         "正常系: 空のソースファイルでもURLあり",
			urlFlags:     []string{"http://example.com/feed.xml"},
			sourceFlag:   "empty_source.txt",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestRecommendCmd()

			for _, url := range tt.urlFlags {
				cmd.Flags().Set("url", url)
			}
			if tt.sourceFlag != "" {
				// sourceFlagが設定されている場合、一時ソースファイルを作成
				switch tt.sourceFlag {
				case "tmp_source.txt":
					content := "http://example.com/from_file.xml\nhttp://another.com/from_file.xml"
					err := os.WriteFile(tt.sourceFlag, []byte(content), 0644)
					assert.NoError(t, err)
					defer os.Remove(tt.sourceFlag)
				case "empty_source.txt":
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

	// 統合テスト: 複数URL処理
	t.Run("統合テスト: 複数URL処理", func(t *testing.T) {
		cmd := newTestRecommendCmd()

		for _, url := range []string{"https://example1.com/feed.xml", "https://example2.com/feed.xml", "https://example3.com/feed.xml"} {
			cmd.Flags().Set("url", url)
		}

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Len(t, params.URLs, 3)
	})

	// 統合テスト: URLとソースの組み合わせ
	t.Run("統合テスト: URLとソースの組み合わせ", func(t *testing.T) {
		cmd := newTestRecommendCmd()

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
		assert.Len(t, params.URLs, 3) // URLから1件 + ソースから2件
	})

	// 正常系: 空行や空白を含むソースファイル
	t.Run("正常系: 空行や空白を含むソースファイル", func(t *testing.T) {
		cmd := newTestRecommendCmd()

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
			assert.True(t, len(url) > 0, "URLは空であってはならない")
		}
	})

	// 正常系: 単一URL指定時の動作確認
	t.Run("正常系: 単一URL指定時の動作確認", func(t *testing.T) {
		cmd := newTestRecommendCmd()

		cmd.Flags().Set("url", "https://example.com/feed.xml")

		params, err := newRecommendParams(cmd)
		assert.NoError(t, err)
		assert.NotNil(t, params)
		assert.Equal(t, []string{"https://example.com/feed.xml"}, params.URLs)
	})
}
