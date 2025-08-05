package infra

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

