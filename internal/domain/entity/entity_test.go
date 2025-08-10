package entity

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestArticle_Validate(t *testing.T) {
	tests := []struct {
		name    string
		article *Article
		wantErr bool
		errors  []string
	}{
		{
			name: "正常系_すべてのフィールドが適切に設定されている",
			article: &Article{
				Title:     "テスト記事",
				Link:      "https://example.com/article",
				Published: &time.Time{},
				Content:   "記事の内容",
			},
			wantErr: false,
			errors:  []string{},
		},
		{
			name: "異常系_Titleが空文字列",
			article: &Article{
				Title:     "",
				Link:      "https://example.com/article",
				Published: &time.Time{},
				Content:   "記事の内容",
			},
			wantErr: true,
			errors:  []string{"記事のタイトルが設定されていません"},
		},
		{
			name: "異常系_Linkが空文字列",
			article: &Article{
				Title:     "テスト記事",
				Link:      "",
				Published: &time.Time{},
				Content:   "記事の内容",
			},
			wantErr: true,
			errors:  []string{"記事のリンクが設定されていません"},
		},
		{
			name: "異常系_LinkがURL形式でない",
			article: &Article{
				Title:     "テスト記事",
				Link:      "invalid-url",
				Published: &time.Time{},
				Content:   "記事の内容",
			},
			wantErr: true,
			errors:  []string{"記事のリンクが正しいURL形式ではありません"},
		},
		{
			name: "異常系_Publishedがnil",
			article: &Article{
				Title:     "テスト記事",
				Link:      "https://example.com/article",
				Published: nil,
				Content:   "記事の内容",
			},
			wantErr: true,
			errors:  []string{"記事の公開日時が設定されていません"},
		},
		{
			name: "異常系_Contentが空文字列",
			article: &Article{
				Title:     "テスト記事",
				Link:      "https://example.com/article",
				Published: &time.Time{},
				Content:   "",
			},
			wantErr: true,
			errors:  []string{"記事の内容が設定されていません"},
		},
		{
			name: "異常系_複数のフィールドでエラー",
			article: &Article{
				Title:     "",
				Link:      "",
				Published: nil,
				Content:   "",
			},
			wantErr: true,
			errors: []string{
				"記事のタイトルが設定されていません",
				"記事のリンクが設定されていません",
				"記事の公開日時が設定されていません",
				"記事の内容が設定されていません",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.article.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)
			if tt.wantErr {
				assert.Equal(t, tt.errors, result.Errors)
			} else {
				assert.Empty(t, result.Errors)
			}
		})
	}
}

func TestRecommend_Validate(t *testing.T) {
	validArticle := Article{
		Title:     "テスト記事",
		Link:      "https://example.com/article",
		Published: &time.Time{},
		Content:   "記事の内容",
	}

	invalidArticle := Article{
		Title:     "",
		Link:      "",
		Published: nil,
		Content:   "",
	}

	tests := []struct {
		name      string
		recommend *Recommend
		wantErr   bool
		hasWarn   bool
	}{
		{
			name: "正常系_有効な記事とコメント",
			recommend: &Recommend{
				Article: validArticle,
				Comment: func() *string { s := "推薦コメント"; return &s }(),
			},
			wantErr: false,
			hasWarn: false,
		},
		{
			name: "正常系_有効な記事でコメントなし",
			recommend: &Recommend{
				Article: validArticle,
				Comment: nil,
			},
			wantErr: false,
			hasWarn: false,
		},
		{
			name: "警告_有効な記事で空のコメント",
			recommend: &Recommend{
				Article: validArticle,
				Comment: func() *string { s := ""; return &s }(),
			},
			wantErr: false,
			hasWarn: true,
		},
		{
			name: "異常系_無効な記事",
			recommend: &Recommend{
				Article: invalidArticle,
				Comment: func() *string { s := "推薦コメント"; return &s }(),
			},
			wantErr: true,
			hasWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.recommend.Validate()

			assert.Equal(t, !tt.wantErr, result.IsValid)

			if tt.hasWarn {
				assert.Contains(t, result.Warnings, "推薦コメントが空です")
			}
		})
	}
}
