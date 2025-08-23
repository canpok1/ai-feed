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
			expectedErr:  "ソースファイルにURLが含まれていません",
		},
		{
			name:         "Empty source file but URLs provided",
			urlFlags:     []string{"http://example.com/feed.xml"},
			sourceFlag:   "empty_source.txt",
			expectedURLs: nil,
			expectedErr:  "ソースファイルにURLが含まれていません",
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
