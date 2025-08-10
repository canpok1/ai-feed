package entity

import (
	"strings"
	"testing"
	"time"
)

func TestPromptConfig_BuildCommentPrompt(t *testing.T) {
	tests := []struct {
		name     string
		template string
		article  *Article
		want     string
	}{
		{
			name:     "新形式のテンプレート（.付き）",
			template: "タイトル: {{.Title}}\nURL: {{.Link}}\n内容: {{.Content}}",
			article: &Article{
				Title:   "テスト記事",
				Link:    "https://example.com",
				Content: "これはテスト内容です",
			},
			want: "タイトル: テスト記事\nURL: https://example.com\n内容: これはテスト内容です",
		},
		{
			name:     "条件分岐を含むテンプレート",
			template: "{{if .Title}}タイトル: {{.Title}}{{end}}\n{{if .Link}}URL: {{.Link}}{{end}}",
			article: &Article{
				Title: "テスト記事",
				Link:  "https://example.com",
			},
			want: "タイトル: テスト記事\nURL: https://example.com",
		},
		{
			name:     "Publishedフィールドへのアクセス",
			template: "タイトル: {{.Title}}{{if .Published}}\n公開日時あり{{end}}",
			article: &Article{
				Title:     "テスト記事",
				Published: &[]time.Time{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}[0],
			},
			want: "タイトル: テスト記事\n公開日時あり",
		},
		{
			name:     "空のフィールドの処理",
			template: "{{if .Title}}タイトル: {{.Title}}{{else}}タイトルなし{{end}}",
			article: &Article{
				Title: "",
			},
			want: "タイトルなし",
		},
		{
			name: "複雑なテンプレート",
			template: `以下の記事の紹介文を100字以内で作成してください。
---
記事タイトル: {{.Title}}
記事URL: {{.Link}}
記事内容:
{{.Content}}`,
			article: &Article{
				Title:   "AIに関する最新記事",
				Link:    "https://example.com/ai-article",
				Content: "AIの進化について詳しく解説しています。",
			},
			want: `以下の記事の紹介文を100字以内で作成してください。
---
記事タイトル: AIに関する最新記事
記事URL: https://example.com/ai-article
記事内容:
AIの進化について詳しく解説しています。`,
		},
		{
			name:     "存在しないキーを持つテンプレート実行エラー時のフォールバック",
			template: "タイトル: {{.Title}}\n{{.Invalid}}",
			article: &Article{
				Title: "テスト記事",
			},
			// 実行エラーの場合、元のテンプレートをそのまま返す
			want: "タイトル: {{.Title}}\n{{.Invalid}}",
		},
		{
			name:     "旧形式のテンプレート構文（後方互換性）",
			template: "タイトル: {{title}}\nURL: {{url}}\n内容: {{content}}",
			article: &Article{
				Title:   "テスト記事",
				Link:    "https://example.com",
				Content: "これはテスト内容です",
			},
			// 旧形式は自動的に新形式に変換される
			want: "タイトル: テスト記事\nURL: https://example.com\n内容: これはテスト内容です",
		},
		{
			name:     "テンプレートパースエラー時のフォールバック",
			template: "タイトル: {{.Title}}\n{{unclosed",
			article: &Article{
				Title: "テスト記事",
			},
			// パースエラーの場合、元のテンプレートをそのまま返す
			want: "タイトル: {{.Title}}\n{{unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &PromptConfig{
				CommentPromptTemplate: tt.template,
			}
			got := c.BuildCommentPrompt(tt.article)
			if got != tt.want {
				t.Errorf("BuildCommentPrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPromptConfig_BuildCommentPrompt_Performance(t *testing.T) {
	// パフォーマンステスト: 大量のコンテンツでも正しく動作することを確認
	longContent := strings.Repeat("これは非常に長いコンテンツです。", 100)

	c := &PromptConfig{
		CommentPromptTemplate: "タイトル: {{.Title}}\n内容: {{.Content}}",
	}

	article := &Article{
		Title:   "パフォーマンステスト",
		Content: longContent,
	}

	result := c.BuildCommentPrompt(article)

	if !strings.Contains(result, "パフォーマンステスト") {
		t.Error("Title not found in result")
	}

	if !strings.Contains(result, longContent) {
		t.Error("Content not found in result")
	}
}

func TestPromptConfig_BuildCommentPrompt_Cache(t *testing.T) {
	// キャッシュのパフォーマンステスト
	template := "タイトル: {{.Title}}\nURL: {{.Link}}\n内容: {{.Content}}"
	c := &PromptConfig{
		CommentPromptTemplate: template,
	}

	article := &Article{
		Title:   "キャッシュテスト",
		Link:    "https://example.com",
		Content: "これはキャッシュテスト内容です",
	}

	expected := "タイトル: キャッシュテスト\nURL: https://example.com\n内容: これはキャッシュテスト内容です"

	// 1回目の呼び出し（キャッシュなし）
	result1 := c.BuildCommentPrompt(article)
	if result1 != expected {
		t.Errorf("First call failed: got %v, want %v", result1, expected)
	}

	// 2回目の呼び出し（キャッシュあり）
	result2 := c.BuildCommentPrompt(article)
	if result2 != expected {
		t.Errorf("Second call failed: got %v, want %v", result2, expected)
	}

	// 結果が同じであることを確認
	if result1 != result2 {
		t.Error("Cache results differ from non-cache results")
	}
}
