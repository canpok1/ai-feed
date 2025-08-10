package entity

import (
	"time"
)

// Article represents a single article in a feed.
type Article struct {
	Title     string
	Link      string
	Published *time.Time
	Content   string
}

// Validate はArticleの内容をバリデーションする
func (a *Article) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Title: 必須項目（空文字列でない）
	if err := ValidateRequired(a.Title, "記事のタイトル"); err != nil {
		builder.AddError(err.Error())
	}

	// Link: 必須項目（空文字列でない）、URL形式であること
	if err := ValidateURL(a.Link, "記事のリンク"); err != nil {
		builder.AddError(err.Error())
	}

	// Published: nilでないこと
	if a.Published == nil {
		builder.AddError("記事の公開日時が設定されていません")
	}

	// Content: 必須項目（空文字列でない）
	if err := ValidateRequired(a.Content, "記事の内容"); err != nil {
		builder.AddError(err.Error())
	}

	return builder.Build()
}

type Recommend struct {
	Article Article
	Comment *string
}

// Validate はRecommendの内容をバリデーションする
func (r *Recommend) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Article: 必須項目、Articleのバリデーションも実行
	builder.MergeResult(r.Article.Validate())

	// Comment: 任意項目だが、設定されている場合は空文字列でないこと
	if r.Comment != nil && *r.Comment == "" {
		builder.AddWarning("推薦コメントが空です")
	}

	return builder.Build()
}
