package entity

import (
	"net/url"
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
	var errors []string

	// Title: 必須項目（空文字列でない）
	if a.Title == "" {
		errors = append(errors, "記事のタイトルが設定されていません")
	}

	// Link: 必須項目（空文字列でない）、URL形式であること
	if a.Link == "" {
		errors = append(errors, "記事のリンクが設定されていません")
	} else {
		// URL形式チェック（絶対URLであることを確認）
		if parsedURL, err := url.Parse(a.Link); err != nil || !parsedURL.IsAbs() {
			errors = append(errors, "記事のリンクが正しいURL形式ではありません")
		}
	}

	// Published: nilでないこと
	if a.Published == nil {
		errors = append(errors, "記事の公開日時が設定されていません")
	}

	// Content: 必須項目（空文字列でない）
	if a.Content == "" {
		errors = append(errors, "記事の内容が設定されていません")
	}

	return &ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

type Recommend struct {
	Article Article
	Comment *string
}

// Validate はRecommendの内容をバリデーションする
func (r *Recommend) Validate() *ValidationResult {
	var errors []string
	var warnings []string

	// Article: 必須項目、Articleのバリデーションも実行
	articleResult := r.Article.Validate()
	if !articleResult.IsValid {
		errors = append(errors, articleResult.Errors...)
	}
	warnings = append(warnings, articleResult.Warnings...)

	// Comment: 任意項目だが、設定されている場合は空文字列でないこと
	if r.Comment != nil && *r.Comment == "" {
		warnings = append(warnings, "推薦コメントが空です")
	}

	return &ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}
