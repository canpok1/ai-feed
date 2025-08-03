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
		urlFlag      string
		sourceFlag   string
		expectedURLs []string
		expectedErr  string
	}{
		{
			name:         "URL flag only",
			urlFlag:      "http://example.com/feed.xml",
			sourceFlag:   "",
			expectedURLs: []string{"http://example.com/feed.xml"},
			expectedErr:  "",
		},
		{
			name:         "Source flag only with valid file",
			urlFlag:      "",
			sourceFlag:   "tmp_source.txt",
			expectedURLs: []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml"},
		},
		{
			name:         "Both URL and source flags",
			urlFlag:      "http://example.com/feed.xml",
			sourceFlag:   "tmp_source.txt",
			expectedURLs: nil,
			expectedErr:  "cannot use --url and --source options together",
		},
		{
			name:         "Neither URL nor source flags",
			urlFlag:      "",
			sourceFlag:   "",
			expectedURLs: nil,
			expectedErr:  "either --url or --source must be specified",
		},
		{
			name:         "Source file not found",
			urlFlag:      "",
			sourceFlag:   "non_existent_file.txt",
			expectedURLs: nil,
			expectedErr:  "failed to read URLs from file: open non_existent_file.txt: no such file or directory",
		},
		{
			name:         "Empty source file",
			urlFlag:      "",
			sourceFlag:   "empty_source.txt",
			expectedURLs: nil,
			expectedErr:  "source file contains no URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy cobra.Command and set flags
			cmd := &cobra.Command{}
			cmd.Flags().StringP("url", "u", "", "URL of the feed to recommend from")
			cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs")

			if tt.urlFlag != "" {
				cmd.Flags().Set("url", tt.urlFlag)
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
