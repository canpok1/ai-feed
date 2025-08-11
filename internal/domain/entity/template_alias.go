package entity

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// TemplateAliasConverter はテンプレートの別名を既存記法に変換する構造体
type TemplateAliasConverter struct {
	// aliasMap は別名から既存記法へのマッピング
	aliasMap map[string]string
}

// NewPromptTemplateAliasConverter はPromptConfig用の別名変換器を作成する
func NewPromptTemplateAliasConverter() *TemplateAliasConverter {
	return &TemplateAliasConverter{
		aliasMap: map[string]string{
			"TITLE":   ".Title",
			"URL":     ".Link",
			"CONTENT": ".Content",
		},
	}
}

// NewSlackTemplateAliasConverter はSlackAPIConfig用の別名変換器を作成する
func NewSlackTemplateAliasConverter() *TemplateAliasConverter {
	return &TemplateAliasConverter{
		aliasMap: map[string]string{
			"TITLE":         ".Article.Title",
			"URL":           ".Article.Link",
			"CONTENT":       ".Article.Content",
			"COMMENT":       ".Comment",
			"FIXED_MESSAGE": ".FixedMessage",
		},
	}
}

// NewMisskeyTemplateAliasConverter はMisskeyConfig用の別名変換器を作成する
func NewMisskeyTemplateAliasConverter() *TemplateAliasConverter {
	return &TemplateAliasConverter{
		aliasMap: map[string]string{
			"TITLE":         ".Article.Title",
			"URL":           ".Article.Link",
			"CONTENT":       ".Article.Content",
			"COMMENT":       ".Comment",
			"FIXED_MESSAGE": ".FixedMessage",
		},
	}
}

// TemplateAliasError はテンプレート別名変換のエラー
type TemplateAliasError struct {
	InvalidAlias string
	Message      string
}

func (e *TemplateAliasError) Error() string {
	return e.Message
}

// Convert はテンプレート文字列内の別名を既存記法に変換する
func (c *TemplateAliasConverter) Convert(template string) (string, error) {
	// 既存記法（{{.で始まるもの）はそのまま通す
	// 別名記法のみを処理する

	// まず、不正な別名記法（小文字を含む、存在しないパラメータ）をチェック
	invalidPattern := regexp.MustCompile(`\{\{([A-Za-z_]+)\}\}`)
	matches := invalidPattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		// 完全な別名記法の部分
		fullMatch := match[0]
		// 中身の部分（{{と}}を除いた部分）
		innerContent := match[1]

		// ドットで始まる場合は既存記法なのでスキップ
		if strings.HasPrefix(fullMatch, "{{.") {
			continue
		}

		// if、end、elseなどのテンプレート制御構文はスキップ
		if isTemplateControl(innerContent) {
			continue
		}

		// 大文字とアンダースコアのみで構成されているかチェック
		if !isUpperCaseOnly(innerContent) {
			// 小文字が含まれている場合はエラー
			suggestion := strings.ToUpper(innerContent)
			if _, exists := c.aliasMap[suggestion]; exists {
				return "", &TemplateAliasError{
					InvalidAlias: fullMatch,
					Message:      fmt.Sprintf("別名記法では大文字のみが許可されています。'%s' の代わりに '{{%s}}' を使用してください", fullMatch, suggestion),
				}
			}
			return "", &TemplateAliasError{
				InvalidAlias: fullMatch,
				Message:      fmt.Sprintf("別名記法では大文字のみが許可されています: '%s'", fullMatch),
			}
		}

		// 存在するパラメータかチェック
		if _, exists := c.aliasMap[innerContent]; !exists {
			// 存在しないパラメータの場合はエラー
			validAliases := c.getValidAliases()
			return "", &TemplateAliasError{
				InvalidAlias: fullMatch,
				Message:      fmt.Sprintf("存在しないパラメータです: '%s'。使用可能なパラメータ: %s", fullMatch, strings.Join(validAliases, ", ")),
			}
		}
	}

	// エラーチェックが通ったら、別名を既存記法に変換
	result := template
	for alias, replacement := range c.aliasMap {
		pattern := fmt.Sprintf("{{%s}}", alias)
		newPattern := fmt.Sprintf("{{%s}}", replacement)
		result = strings.ReplaceAll(result, pattern, newPattern)
	}

	return result, nil
}

// isUpperCaseOnly は文字列が大文字とアンダースコアのみで構成されているかチェック
func isUpperCaseOnly(s string) bool {
	for _, r := range s {
		if r != '_' && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

// isTemplateControl はテンプレート制御構文かどうかチェック
func isTemplateControl(s string) bool {
	// テンプレート制御構文のキーワード
	controls := []string{"if", "else", "end", "range", "with", "define", "template", "block"}

	// 先頭の単語を取得
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return false
	}

	firstWord := strings.ToLower(parts[0])
	return slices.Contains(controls, firstWord)
}

// getValidAliases は使用可能な別名のリストを返す
func (c *TemplateAliasConverter) getValidAliases() []string {
	aliases := make([]string, 0, len(c.aliasMap))
	for alias := range c.aliasMap {
		aliases = append(aliases, "{{"+alias+"}}")
	}
	return aliases
}
