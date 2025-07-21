package cmd

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/domain/mock_domain"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/spf13/cobra"
)

// createMockConfig creates a mock entity.Config for testing purposes.
func createMockConfig(modelName, promptName string) *entity.Config {
	return &entity.Config{
		General: entity.GeneralConfig{
			DefaultExecutionProfile: "default",
		},
		AIModels: map[string]entity.AIModelConfig{
			modelName: {Type: "test-type", APIKey: "test-key"},
		},
		Prompts: map[string]entity.PromptConfig{
			promptName: {SystemMessage: "test-system-message", CommentPromptTemplate: "test-prompt-template"},
		},
		ExecutionProfiles: map[string]entity.ExecutionProfile{
			"default": {AIModel: modelName, Prompt: promptName},
		},
	}
}

func toStringP(value string) *string {
	return &value
}

func TestInstantRecommendRunner_Run(t *testing.T) {
	tests := []struct {
		name                        string
		mockFetchClientExpectations func(m *mock_domain.MockFetchClient)
		mockRecommenderExpectations func(m *mock_domain.MockRecommender)
		params                      *instantRecommendParams
		expectedStdout              string
		expectedStderr              string
		expectedErrorMessage        *string
	}{
		{
			name: "Successful recommendation",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Return(&entity.Recommend{
					Article: entity.Article{Title: "Recommended Article", Link: "http://example.com/recommended"},
				}, nil).Times(1)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout: strings.Join([]string{
				"Title: Recommended Article",
				"Link: http://example.com/recommended",
				"",
			}, "\n"),
			expectedStderr:       "",
			expectedErrorMessage: nil,
		},
		{
			name: "No articles found",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				// Should not be called if no articles are found
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/empty.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout:       "No articles found in the feed.\n",
			expectedStderr:       "",
			expectedErrorMessage: nil,
		},
		{
			name: "Fetch error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return(nil, fmt.Errorf("mock fetch error")).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://invalid.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout:       "No articles found in the feed.\n",
			expectedStderr:       "Error fetching feed from http://invalid.com/feed.xml: mock fetch error\n",
			expectedErrorMessage: nil,
		},
		{
			name: "Recommend error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {
				m.EXPECT().Fetch(gomock.Any()).Return([]entity.Article{
					{Title: "Test Article", Link: "http://example.com/test"},
				}, nil).Times(1)
			},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {
				m.EXPECT().Recommend(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("mock recommend error")).Times(1)
			},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: createMockConfig("test-model", "test-prompt"),
			},
			expectedStdout:       "",
			expectedStderr:       "",
			expectedErrorMessage: toStringP("failed to recommend article: mock recommend error"),
		},
		{
			name:                        "GetDefaultAIModel error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {},
			params: &instantRecommendParams{
				urls:   []string{"http://example.com/feed.xml"},
				config: &entity.Config{}, // No default AI model set
			},
			expectedStdout:       "",
			expectedErrorMessage: toStringP("failed to get default AI model: default execution profile not found: "),
		},
		{
			name:                        "GetDefaultPrompt error",
			mockFetchClientExpectations: func(m *mock_domain.MockFetchClient) {},
			mockRecommenderExpectations: func(m *mock_domain.MockRecommender) {},
			params: &instantRecommendParams{
				urls: []string{"http://example.com/feed.xml"},
				config: &entity.Config{
					General: entity.GeneralConfig{DefaultExecutionProfile: "default"},
					AIModels: map[string]entity.AIModelConfig{
						"test-model": {Type: "test-type"},
					},
					ExecutionProfiles: map[string]entity.ExecutionProfile{
						"default": {AIModel: "test-model", Prompt: ""}, // No prompt set
					},
				},
			},
			expectedStdout:       "",
			expectedStderr:       "",
			expectedErrorMessage: toStringP("failed to get default prompt: prompt not found: "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockFetchClient := mock_domain.NewMockFetchClient(ctrl)
			tt.mockFetchClientExpectations(mockFetchClient)

			mockRecommender := mock_domain.NewMockRecommender(ctrl)
			tt.mockRecommenderExpectations(mockRecommender)

			stdoutBuffer := new(bytes.Buffer)
			stderrBuffer := new(bytes.Buffer)

			runner, err := newInstantRecommendRunner(mockFetchClient, mockRecommender, stdoutBuffer, stderrBuffer)
			assert.NoError(t, err, "newInstantRecommendRunner should not return an error")

			// Create a dummy cobra.Command for the Run method, as it's required but not directly used in runner.Run logic
			cmd := &cobra.Command{}
			cmd.SetOut(stdoutBuffer)
			cmd.SetErr(stderrBuffer)

			err = runner.Run(cmd, tt.params)
			hasError := err != nil
			expectedHasError := tt.expectedErrorMessage != nil

			assert.Equal(t, expectedHasError, hasError, "Expected error state mismatch")
			if hasError {
				assert.Contains(t, err.Error(), *tt.expectedErrorMessage, "Error message mismatch")
			}

			assert.Equal(t, tt.expectedStdout, stdoutBuffer.String(), "Stdout mismatch")
			assert.Equal(t, tt.expectedStderr, stderrBuffer.String(), "Stderr mismatch")
		})
	}
}

func TestNewInstantRecommendParams(t *testing.T) {
	tests := []struct {
		name           string
		urlFlag        string
		sourceFlag     string
		mockConfig     *entity.Config
		configLoadErr  error
		expectedURLs   []string
		expectedConfig *entity.Config
		expectedErr    string
	}{
		{
			name:           "URL flag only",
			urlFlag:        "http://example.com/feed.xml",
			sourceFlag:     "",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   []string{"http://example.com/feed.xml"},
			expectedConfig: createMockConfig("test-model", "test-prompt"),
			expectedErr:    "",
		},
		{
			name:           "Source flag only with valid file",
			urlFlag:        "",
			sourceFlag:     "tmp_source.txt",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   []string{"http://example.com/from_file.xml", "http://another.com/from_file.xml"},
			expectedConfig: createMockConfig("test-model", "test-prompt"),
			expectedErr:    "",
		},
		{
			name:           "Both URL and source flags",
			urlFlag:        "http://example.com/feed.xml",
			sourceFlag:     "tmp_source.txt",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   nil,
			expectedConfig: nil,
			expectedErr:    "cannot use --url and --source options together",
		},
		{
			name:           "Neither URL nor source flags",
			urlFlag:        "",
			sourceFlag:     "",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   nil,
			expectedConfig: nil,
			expectedErr:    "either --url or --source must be specified",
		},
		{
			name:           "Source file not found",
			urlFlag:        "",
			sourceFlag:     "non_existent_file.txt",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   nil,
			expectedConfig: nil,
			expectedErr:    "failed to read URLs from file: open non_existent_file.txt: no such file or directory",
		},
		{
			name:           "Empty source file",
			urlFlag:        "",
			sourceFlag:     "empty_source.txt",
			mockConfig:     createMockConfig("test-model", "test-prompt"),
			configLoadErr:  nil,
			expectedURLs:   nil,
			expectedConfig: nil,
			expectedErr:    "source file contains no URLs",
		},
		{
			name:           "Config load error",
			urlFlag:        "http://example.com/feed.xml",
			sourceFlag:     "",
			mockConfig:     nil,
			configLoadErr:  fmt.Errorf("mock config load error"),
			expectedURLs:   nil,
			expectedConfig: nil,
			expectedErr:    "failed to load config: mock config load error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Create a mock ConfigRepository
			mockConfigRepo := mock_domain.NewMockConfigRepository(ctrl)
			// Determine if configRepo.Load() should be called
			shouldCallLoad := true
			if (tt.urlFlag != "" && tt.sourceFlag != "") || (tt.urlFlag == "" && tt.sourceFlag == "") || (tt.sourceFlag != "" && strings.Contains(tt.expectedErr, "failed to read URLs from file")) || (tt.sourceFlag != "" && strings.Contains(tt.expectedErr, "source file contains no URLs")) {
				shouldCallLoad = false
			}

			if shouldCallLoad {
				if tt.configLoadErr != nil {
					mockConfigRepo.EXPECT().Load().Return(nil, tt.configLoadErr).Times(1)
				} else {
					mockConfigRepo.EXPECT().Load().Return(tt.mockConfig, nil).Times(1)
				}
			} else {
				mockConfigRepo.EXPECT().Load().Times(0)
			}

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

			params, err := newInstantRecommendParams(cmd, mockConfigRepo)

			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, params)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params)
				assert.Equal(t, tt.expectedURLs, params.urls)
				assert.Equal(t, tt.expectedConfig, params.config)
			}
		})
	}
}
