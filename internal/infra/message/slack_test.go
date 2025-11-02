package message

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSlackClient is a mock of the slackClient interface.
type MockSlackClient struct {
	mock.Mock
}

func (m *MockSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	args := m.Called(channelID, options)
	return args.String(0), args.String(1), args.Error(2)
}

func TestSendRecommend_PostMessageOptions(t *testing.T) {
	// ヘルパー関数: SecretStringを作成
	makeSecretString := func(value string) entity.SecretString {
		var s entity.SecretString
		s.UnmarshalText([]byte(value))
		return s
	}

	tests := []struct {
		name      string
		config    *entity.SlackAPIConfig
		setupMock func(*MockSlackClient, *entity.SlackAPIConfig)
	}{
		{
			name: "no extra options",
			config: &entity.SlackAPIConfig{
				APIToken:        makeSecretString("xoxb-test-token"),
				Channel:         "#test",
				MessageTemplate: testutil.StringPtr("test message"),
			},
			setupMock: func(m *MockSlackClient, c *entity.SlackAPIConfig) {
				m.On("PostMessage", c.Channel, mock.AnythingOfType("[]slack.MsgOption")).Return("", "", nil).Run(func(args mock.Arguments) {
					opts := args.Get(1).([]slack.MsgOption)
					assert.Len(t, opts, 1)
				})
			},
		},
		{
			name: "with username",
			config: &entity.SlackAPIConfig{
				APIToken:        makeSecretString("xoxb-test-token"),
				Channel:         "#test",
				MessageTemplate: testutil.StringPtr("test message"),
				Username:        testutil.StringPtr("test-user"),
			},
			setupMock: func(m *MockSlackClient, c *entity.SlackAPIConfig) {
				m.On("PostMessage", c.Channel, mock.AnythingOfType("[]slack.MsgOption")).Return("", "", nil).Run(func(args mock.Arguments) {
					opts := args.Get(1).([]slack.MsgOption)
					assert.Len(t, opts, 2)
				})
			},
		},
		{
			name: "with icon url",
			config: &entity.SlackAPIConfig{
				APIToken:        makeSecretString("xoxb-test-token"),
				Channel:         "#test",
				MessageTemplate: testutil.StringPtr("test message"),
				IconURL:         testutil.StringPtr("http://example.com/icon.png"),
			},
			setupMock: func(m *MockSlackClient, c *entity.SlackAPIConfig) {
				m.On("PostMessage", c.Channel, mock.AnythingOfType("[]slack.MsgOption")).Return("", "", nil).Run(func(args mock.Arguments) {
					opts := args.Get(1).([]slack.MsgOption)
					assert.Len(t, opts, 2)
				})
			},
		},
		{
			name: "with icon emoji",
			config: &entity.SlackAPIConfig{
				APIToken:        makeSecretString("xoxb-test-token"),
				Channel:         "#test",
				MessageTemplate: testutil.StringPtr("test message"),
				IconEmoji:       testutil.StringPtr(":smile:"),
			},
			setupMock: func(m *MockSlackClient, c *entity.SlackAPIConfig) {
				m.On("PostMessage", c.Channel, mock.AnythingOfType("[]slack.MsgOption")).Return("", "", nil).Run(func(args mock.Arguments) {
					opts := args.Get(1).([]slack.MsgOption)
					assert.Len(t, opts, 2)
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockSlackClient)
			currentTT := tt // capture range variable
			currentTT.setupMock(mockClient, currentTT.config)

			sender := NewSlackSender(tt.config, mockClient)
			recommend := &entity.Recommend{
				Article: entity.Article{Title: "Test", Link: "http://test.com"},
				Comment: testutil.StringPtr("test comment"),
			}

			err := sender.SendRecommend(recommend, "fixed message")
			require.NoError(t, err)
			mockClient.AssertExpectations(t)
		})
	}
}

// TestSlackTemplateExecution はSlackメッセージテンプレートの実行をテストする
func TestSlackTemplateExecution(t *testing.T) {
	publishedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		messageTemplate string
		templateData    *SlackTemplateData
		expectedMessage string
		expectError     bool
	}{
		{
			name:            "デフォルトテンプレート（全フィールドあり）",
			messageTemplate: "{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      testutil.StringPtr("これは推薦コメントです。"),
				FixedMessage: "固定メッセージです。",
			},
			expectedMessage: "これは推薦コメントです。\nテスト記事\nhttps://example.com/article\n固定メッセージです。",
			expectError:     false,
		},
		{
			name:            "デフォルトテンプレート（コメントなし）",
			messageTemplate: "{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      nil,
				FixedMessage: "固定メッセージです。",
			},
			expectedMessage: "テスト記事\nhttps://example.com/article\n固定メッセージです。",
			expectError:     false,
		},
		{
			name:            "デフォルトテンプレート（固定メッセージなし）",
			messageTemplate: "{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      testutil.StringPtr("これは推薦コメントです。"),
				FixedMessage: "",
			},
			expectedMessage: "これは推薦コメントです。\nテスト記事\nhttps://example.com/article",
			expectError:     false,
		},
		{
			name:            "デフォルトテンプレート（コメントと固定メッセージなし）",
			messageTemplate: "{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      nil,
				FixedMessage: "",
			},
			expectedMessage: "テスト記事\nhttps://example.com/article",
			expectError:     false,
		},
		{
			name:            "カスタムテンプレート（全フィールド使用）",
			messageTemplate: "記事: {{.Article.Title}} ({{.Article.Link}}){{if .Comment}}\nコメント: {{.Comment}}{{end}}{{if .FixedMessage}}\n補足: {{.FixedMessage}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "カスタム記事",
					Link:      "https://example.com/custom",
					Published: &publishedTime,
					Content:   "カスタム内容",
				},
				Comment:      testutil.StringPtr("カスタムコメント"),
				FixedMessage: "カスタム固定メッセージ",
			},
			expectedMessage: "記事: カスタム記事 (https://example.com/custom)\nコメント: カスタムコメント\n補足: カスタム固定メッセージ",
			expectError:     false,
		},
		{
			name:            "シンプルなカスタムテンプレート",
			messageTemplate: "{{.Article.Title}} - {{.Article.Link}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "シンプル記事",
					Link:      "https://example.com/simple",
					Published: &publishedTime,
					Content:   "シンプル内容",
				},
				Comment:      testutil.StringPtr("シンプルコメント"),
				FixedMessage: "シンプル固定メッセージ",
			},
			expectedMessage: "シンプル記事 - https://example.com/simple",
			expectError:     false,
		},
		{
			name:            "日本語テンプレート",
			messageTemplate: "タイトル: {{.Article.Title}}\nリンク: {{.Article.Link}}{{if .Comment}}\n推薦理由: {{.Comment}}{{end}}",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "日本語記事タイトル",
					Link:      "https://example.com/japanese-article",
					Published: &publishedTime,
					Content:   "日本語記事内容",
				},
				Comment:      testutil.StringPtr("この記事は非常に有用です。"),
				FixedMessage: "",
			},
			expectedMessage: "タイトル: 日本語記事タイトル\nリンク: https://example.com/japanese-article\n推薦理由: この記事は非常に有用です。",
			expectError:     false,
		},
		{
			name:            "無効なテンプレート構文",
			messageTemplate: "{{.Article.Title",
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "エラー記事",
					Link:      "https://example.com/error",
					Published: &publishedTime,
					Content:   "エラー内容",
				},
				Comment:      testutil.StringPtr("エラーコメント"),
				FixedMessage: "エラー固定メッセージ",
			},
			expectedMessage: "",
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMessage, err := executeSlackTemplate(tt.messageTemplate, tt.templateData)

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid template")
			} else {
				assert.NoError(t, err, "Should not return error for valid template")
				assert.Equal(t, tt.expectedMessage, actualMessage, "Generated message should match expected")
			}
		})
	}
}

// executeSlackTemplate はテスト用のヘルパー関数：Slackテンプレートを実行してメッセージを生成する
func executeSlackTemplate(templateStr string, data *SlackTemplateData) (string, error) {
	tmpl, err := template.New("slack_message").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
