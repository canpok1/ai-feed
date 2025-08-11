package message

import (
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMessageBuilder(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "正常なテンプレート",
			template:    "Title: {{ .Article.Title }}",
			expectError: false,
		},
		{
			name:          "無効なテンプレート",
			template:      "Title: {{ .Article.Title",
			expectError:   true,
			errorContains: "template:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewMessageBuilder(tt.template)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, builder)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, builder)
			}
		})
	}
}

func TestNewMessageBuilder_WithAliases(t *testing.T) {
	tests := []struct {
		name          string
		template      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "別名記法を含むテンプレート",
			template:    "Title: {{TITLE}}\nURL: {{URL}}",
			expectError: false,
		},
		{
			name:        "新旧記法の混在",
			template:    "{{TITLE}} - {{.Article.Link}}",
			expectError: false,
		},
		{
			name:        "条件分岐と別名記法",
			template:    "{{if .Comment}}{{COMMENT}}{{end}}{{TITLE}}",
			expectError: false,
		},
		{
			name:          "小文字の別名記法でエラー",
			template:      "{{title}}",
			expectError:   true,
			errorContains: "別名記法では大文字のみが許可されています",
		},
		{
			name:          "存在しない別名でエラー",
			template:      "{{AUTHOR}}",
			expectError:   true,
			errorContains: "存在しないパラメータです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewMessageBuilder(tt.template)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, builder)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, builder)
			}
		})
	}
}

func TestMessageBuilder_BuildRecommendMessage(t *testing.T) {
	now := time.Now()
	comment := "おすすめの記事です"

	tests := []struct {
		name         string
		template     string
		recommend    *entity.Recommend
		fixedMessage string
		expected     string
		expectError  bool
	}{
		{
			name:     "全フィールドあり",
			template: "Title: {{ .Article.Title }}\nLink: {{ .Article.Link }}\n{{ if .Comment }}Comment: {{ .Comment }}\n{{ end }}{{ if .FixedMessage }}Fixed: {{ .FixedMessage }}{{ end }}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com",
					Published: &now,
					Content:   "テスト内容",
				},
				Comment: &comment,
			},
			fixedMessage: "固定メッセージ",
			expected:     "Title: テスト記事\nLink: https://example.com\nComment: おすすめの記事です\nFixed: 固定メッセージ",
		},
		{
			name:     "CommentとFixedMessageがnil/空",
			template: "Title: {{ .Article.Title }}\n{{ if .Comment }}Comment: {{ .Comment }}\n{{ end }}{{ if .FixedMessage }}Fixed: {{ .FixedMessage }}\n{{ end }}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title: "テスト記事",
					Link:  "https://example.com",
				},
				Comment: nil,
			},
			fixedMessage: "",
			expected:     "Title: テスト記事\n",
		},
		{
			name:     "Publishedフィールドの処理",
			template: "{{ if .Article.Published }}Published: {{ .Article.Published.Format \"2006-01-02\" }}\n{{ end }}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:     "テスト記事",
					Published: &now,
				},
			},
			expected: "Published: " + now.Format("2006-01-02") + "\n",
		},
		{
			name:     "Publishedがnil",
			template: "{{ if .Article.Published }}Published: {{ .Article.Published.Format \"2006-01-02\" }}\n{{ else }}No date\n{{ end }}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:     "テスト記事",
					Published: nil,
				},
			},
			expected: "No date\n",
		},
		{
			name:         "recommendがnil",
			template:     "Title: {{ .Article.Title }}",
			recommend:    nil,
			fixedMessage: "テストメッセージ",
			expected:     "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewMessageBuilder(tt.template)
			require.NoError(t, err)

			result, err := builder.BuildRecommendMessage(tt.recommend, tt.fixedMessage)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageBuilder_BuildRecommendMessage_WithAliases(t *testing.T) {
	now := time.Now()
	comment := "おすすめの記事です"

	tests := []struct {
		name         string
		template     string
		recommend    *entity.Recommend
		fixedMessage string
		expected     string
		expectError  bool
	}{
		{
			name:     "別名記法のみのテンプレート",
			template: "{{TITLE}}\n{{URL}}\n{{if .Comment}}{{COMMENT}}{{end}}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:     "テスト記事",
					Link:      "https://example.com",
					Published: &now,
					Content:   "テスト内容",
				},
				Comment: &comment,
			},
			expected: "テスト記事\nhttps://example.com\nおすすめの記事です",
		},
		{
			name:     "新旧記法の混在テンプレート",
			template: "Title: {{TITLE}}\nURL: {{.Article.Link}}\n{{if .Comment}}Comment: {{COMMENT}}{{end}}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:   "混在テスト",
					Link:    "https://test.com",
					Content: "内容",
				},
				Comment: &comment,
			},
			expected: "Title: 混在テスト\nURL: https://test.com\nComment: おすすめの記事です",
		},
		{
			name:     "FIXED_MESSAGE別名の使用",
			template: "{{TITLE}}\n{{if .FixedMessage}}{{FIXED_MESSAGE}}{{end}}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title: "固定メッセージテスト",
				},
			},
			fixedMessage: "重要なお知らせ",
			expected:     "固定メッセージテスト\n重要なお知らせ",
		},
		{
			name:     "全別名記法の使用",
			template: "{{TITLE}}\n{{URL}}\n{{CONTENT}}\n{{if .Comment}}{{COMMENT}}\n{{end}}{{if .FixedMessage}}{{FIXED_MESSAGE}}{{end}}",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title:   "フルテスト",
					Link:    "https://full.test",
					Content: "完全なテスト内容",
				},
				Comment: &comment,
			},
			fixedMessage: "追加メッセージ",
			expected:     "フルテスト\nhttps://full.test\n完全なテスト内容\nおすすめの記事です\n追加メッセージ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewMessageBuilder(tt.template)
			require.NoError(t, err)

			result, err := builder.BuildRecommendMessage(tt.recommend, tt.fixedMessage)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestMessageBuilder_BuildRecommendMessage_ExecutionError(t *testing.T) {
	// 実行時エラーのテスト（存在しないフィールドへのアクセス）
	template := "{{ .Article.NonExistentField }}"
	builder, err := NewMessageBuilder(template)
	require.NoError(t, err)

	recommend := &entity.Recommend{
		Article: entity.Article{
			Title: "テスト",
		},
	}

	_, err = builder.BuildRecommendMessage(recommend, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NonExistentField")
}
