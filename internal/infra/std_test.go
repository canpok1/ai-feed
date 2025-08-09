package infra

import (
	"bytes"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStdViewer_SendRecommend(t *testing.T) {
	comment := "これはおすすめの記事です"
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		recommend    *entity.Recommend
		fixedMessage string
		expected     string
	}{
		{
			name: "全フィールドあり",
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
			expected:     "Title: テスト記事\nLink: https://example.com\nComment: これはおすすめの記事です\nFixed Message: 固定メッセージ\n",
		},
		{
			name: "Commentがnil",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title: "テスト記事",
					Link:  "https://example.com",
				},
				Comment: nil,
			},
			fixedMessage: "固定メッセージ",
			expected:     "Title: テスト記事\nLink: https://example.com\nFixed Message: 固定メッセージ\n",
		},
		{
			name: "FixedMessageが空",
			recommend: &entity.Recommend{
				Article: entity.Article{
					Title: "テスト記事",
					Link:  "https://example.com",
				},
				Comment: &comment,
			},
			fixedMessage: "",
			expected:     "Title: テスト記事\nLink: https://example.com\nComment: これはおすすめの記事です\n",
		},
		{
			name:         "recommendがnil",
			recommend:    nil,
			fixedMessage: "固定メッセージ",
			expected:     "No articles found in the feed.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			viewer, err := NewStdViewer(&buf)
			require.NoError(t, err)

			err = viewer.SendRecommend(tt.recommend, tt.fixedMessage)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
