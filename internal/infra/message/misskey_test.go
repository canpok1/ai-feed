package message

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMisskeySender はNewMisskeySender関数をテストする
func TestNewMisskeySender(t *testing.T) {
	tests := []struct {
		name             string
		instanceURL      string
		accessToken      string
		messageTemplate  *string
		expectedTemplate string
		expectError      bool
	}{
		{
			name:             "カスタムテンプレート",
			instanceURL:      "https://misskey.example.com",
			accessToken:      "test-token",
			messageTemplate:  testutil.StringPtr("タイトル: {{.Article.Title}}\nURL: {{.Article.Link}}"),
			expectedTemplate: "タイトル: {{.Article.Title}}\nURL: {{.Article.Link}}",
			expectError:      false,
		},
		{
			name:            "無効なURL",
			instanceURL:     "invalid-url",
			accessToken:     "test-token",
			messageTemplate: nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viewer, err := NewMisskeySender(tt.instanceURL, tt.accessToken, tt.messageTemplate)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, viewer)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, viewer)

			misskeySender, ok := viewer.(*MisskeySender)
			require.True(t, ok, "Should be MisskeySender type")

			// テンプレートが正しくパースされていることを確認
			require.NotNil(t, misskeySender.tmpl, "Template should be parsed and stored")

			// テンプレートの内容を確認するため、空のデータで実行してみる
			var buf bytes.Buffer
			testData := &MisskeyTemplateData{
				Article: &entity.Article{
					Title: "Test Title",
					Link:  "https://test.com",
				},
			}
			err = misskeySender.tmpl.Execute(&buf, testData)
			assert.NoError(t, err, "Template should be executable")
			assert.NotEmpty(t, buf.String(), "Template execution should produce output")

			assert.NotNil(t, misskeySender.client)
		})
	}
}

// TestMisskeyTemplateExecution はMisskeyメッセージテンプレートの実行をテストする
func TestMisskeyTemplateExecution(t *testing.T) {
	publishedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		messageTemplate string
		templateData    *MisskeyTemplateData
		expectedMessage string
		expectError     bool
	}{
		{
			name:            "デフォルトテンプレート（全フィールドあり）",
			messageTemplate: "{{if .Comment}}{{.Comment}}\n{{end}}{{.Article.Title}}\n{{.Article.Link}}{{if .FixedMessage}}\n{{.FixedMessage}}{{end}}",
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			templateData: &MisskeyTemplateData{
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
			actualMessage, err := executeMisskeyTemplate(tt.messageTemplate, tt.templateData)

			if tt.expectError {
				assert.Error(t, err, "Should return error for invalid template")
			} else {
				assert.NoError(t, err, "Should not return error for valid template")
				assert.Equal(t, tt.expectedMessage, actualMessage, "Generated message should match expected")
			}
		})
	}
}

// executeMisskeyTemplate はテスト用のヘルパー関数：Misskeyテンプレートを実行してメッセージを生成する
func executeMisskeyTemplate(templateStr string, data *MisskeyTemplateData) (string, error) {
	tmpl, err := template.New("misskey_message").Parse(templateStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
