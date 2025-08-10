package message

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSlackSender はNewSlackSender関数をテストする
func TestNewSlackSender(t *testing.T) {
	tests := []struct {
		name             string
		config           *entity.SlackAPIConfig
		expectedTemplate string
	}{
		{
			name: "デフォルトテンプレート（MessageTemplateがnil）",
			config: &entity.SlackAPIConfig{
				APIToken:        "xoxb-test-token",
				Channel:         "#test",
				MessageTemplate: nil,
			},
			expectedTemplate: DefaultSlackMessageTemplate,
		},
		{
			name: "デフォルトテンプレート（MessageTemplateが空文字列）",
			config: &entity.SlackAPIConfig{
				APIToken:        "xoxb-test-token",
				Channel:         "#test",
				MessageTemplate: stringPtr(""),
			},
			expectedTemplate: DefaultSlackMessageTemplate,
		},
		{
			name: "デフォルトテンプレート（MessageTemplateが空白のみ）",
			config: &entity.SlackAPIConfig{
				APIToken:        "xoxb-test-token",
				Channel:         "#test",
				MessageTemplate: stringPtr("   \n\t  "),
			},
			expectedTemplate: DefaultSlackMessageTemplate,
		},
		{
			name: "カスタムテンプレート",
			config: &entity.SlackAPIConfig{
				APIToken:        "xoxb-test-token",
				Channel:         "#test",
				MessageTemplate: stringPtr("タイトル: {{.Article.Title}}\nURL: {{.Article.Link}}"),
			},
			expectedTemplate: "タイトル: {{.Article.Title}}\nURL: {{.Article.Link}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewer := NewSlackSender(tt.config)
			require.NotNil(t, viewer)

			slackSender, ok := viewer.(*SlackSender)
			require.True(t, ok, "Should be SlackSender type")

			// テンプレートが正しくパースされていることを確認
			require.NotNil(t, slackSender.tmpl, "Template should be parsed and stored")

			// テンプレートの内容を確認するため、空のデータで実行してみる
			var buf bytes.Buffer
			testData := &SlackTemplateData{
				Article: &entity.Article{
					Title: "Test Title",
					Link:  "https://test.com",
				},
			}
			err := slackSender.tmpl.Execute(&buf, testData)
			assert.NoError(t, err, "Template should be executable")
			assert.NotEmpty(t, buf.String(), "Template execution should produce output")

			assert.Equal(t, tt.config.Channel, slackSender.channelID)
			assert.NotNil(t, slackSender.client)
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
			messageTemplate: DefaultSlackMessageTemplate,
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      stringPtr("これは推薦コメントです。"),
				FixedMessage: "固定メッセージです。",
			},
			expectedMessage: "これは推薦コメントです。\nテスト記事\nhttps://example.com/article\n固定メッセージです。",
			expectError:     false,
		},
		{
			name:            "デフォルトテンプレート（コメントなし）",
			messageTemplate: DefaultSlackMessageTemplate,
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
			messageTemplate: DefaultSlackMessageTemplate,
			templateData: &SlackTemplateData{
				Article: &entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com/article",
					Published: &publishedTime,
					Content:   "記事の内容",
				},
				Comment:      stringPtr("これは推薦コメントです。"),
				FixedMessage: "",
			},
			expectedMessage: "これは推薦コメントです。\nテスト記事\nhttps://example.com/article",
			expectError:     false,
		},
		{
			name:            "デフォルトテンプレート（コメントと固定メッセージなし）",
			messageTemplate: DefaultSlackMessageTemplate,
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
				Comment:      stringPtr("カスタムコメント"),
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
				Comment:      stringPtr("シンプルコメント"),
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
				Comment:      stringPtr("この記事は非常に有用です。"),
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
				Comment:      stringPtr("エラーコメント"),
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

// stringPtr はstring値のポインタを返すヘルパー関数
func stringPtr(s string) *string {
	return &s
}
