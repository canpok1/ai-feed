package entity

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPromptTemplateAliasConverter(t *testing.T) {
	converter := NewPromptTemplateAliasConverter()
	assert.NotNil(t, converter)
	assert.NotNil(t, converter.aliasMap)
	assert.Equal(t, 3, len(converter.aliasMap))
	assert.Equal(t, ".Title", converter.aliasMap["TITLE"])
	assert.Equal(t, ".Link", converter.aliasMap["URL"])
	assert.Equal(t, ".Content", converter.aliasMap["CONTENT"])
}

func TestNewSlackTemplateAliasConverter(t *testing.T) {
	converter := NewSlackTemplateAliasConverter()
	assert.NotNil(t, converter)
	assert.NotNil(t, converter.aliasMap)
	assert.Equal(t, 5, len(converter.aliasMap))
	assert.Equal(t, ".Article.Title", converter.aliasMap["TITLE"])
	assert.Equal(t, ".Article.Link", converter.aliasMap["URL"])
	assert.Equal(t, ".Article.Content", converter.aliasMap["CONTENT"])
	assert.Equal(t, ".Comment", converter.aliasMap["COMMENT"])
	assert.Equal(t, ".FixedMessage", converter.aliasMap["FIXED_MESSAGE"])
}

func TestPromptTemplateAliasConverter_Convert(t *testing.T) {
	converter := NewPromptTemplateAliasConverter()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorMsg    string
	}{
		// 正常系テスト
		{
			name:        "単一の別名記法",
			input:       "記事タイトル: {{TITLE}}",
			expected:    "記事タイトル: {{.Title}}",
			expectError: false,
		},
		{
			name:        "複数の別名記法",
			input:       "{{TITLE}} - {{URL}} - {{CONTENT}}",
			expected:    "{{.Title}} - {{.Link}} - {{.Content}}",
			expectError: false,
		},
		{
			name:        "新旧記法の混在",
			input:       "{{TITLE}} - {{.Link}} - {{CONTENT}}",
			expected:    "{{.Title}} - {{.Link}} - {{.Content}}",
			expectError: false,
		},
		{
			name:        "既存記法のみ",
			input:       "{{.Title}} - {{.Link}} - {{.Content}}",
			expected:    "{{.Title}} - {{.Link}} - {{.Content}}",
			expectError: false,
		},
		{
			name:        "テンプレート制御構文との共存",
			input:       "{{if .Title}}{{TITLE}}{{else}}No title{{end}}",
			expected:    "{{if .Title}}{{.Title}}{{else}}No title{{end}}",
			expectError: false,
		},
		{
			name:        "アンダースコアを含む別名",
			input:       "{{FIXED_MESSAGE}}",
			expected:    "{{FIXED_MESSAGE}}",
			expectError: true,
			errorMsg:    "存在しないパラメータです: '{{FIXED_MESSAGE}}'",
		},

		// 異常系テスト
		{
			name:        "小文字の別名記法",
			input:       "{{title}}",
			expected:    "",
			expectError: true,
			errorMsg:    "別名記法では大文字のみが許可されています。'{{title}}' の代わりに '{{TITLE}}' を使用してください",
		},
		{
			name:        "大文字小文字混在",
			input:       "{{Title}}",
			expected:    "",
			expectError: true,
			errorMsg:    "別名記法では大文字のみが許可されています。'{{Title}}' の代わりに '{{TITLE}}' を使用してください",
		},
		{
			name:        "存在しないパラメータ",
			input:       "{{INVALID}}",
			expected:    "",
			expectError: true,
			errorMsg:    "存在しないパラメータです: '{{INVALID}}'",
		},
		{
			name:        "ドット記法混在エラー",
			input:       "{{.TITLE}}",
			expected:    "{{.TITLE}}",
			expectError: false, // 既存記法として扱われるためエラーにならない
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if aliasErr, ok := err.(*TemplateAliasError); ok {
					assert.Contains(t, aliasErr.Message, tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSlackTemplateAliasConverter_Convert(t *testing.T) {
	converter := NewSlackTemplateAliasConverter()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorMsg    string
	}{
		// 正常系テスト
		{
			name:        "Slack用の全別名記法",
			input:       "{{TITLE}} {{URL}} {{CONTENT}} {{COMMENT}} {{FIXED_MESSAGE}}",
			expected:    "{{.Article.Title}} {{.Article.Link}} {{.Article.Content}} {{.Comment}} {{.FixedMessage}}",
			expectError: false,
		},
		{
			name:        "新旧記法の混在",
			input:       "{{TITLE}} - {{.Article.Link}} - {{COMMENT}}",
			expected:    "{{.Article.Title}} - {{.Article.Link}} - {{.Comment}}",
			expectError: false,
		},
		{
			name:        "条件分岐との組み合わせ",
			input:       "{{if .Comment}}{{COMMENT}}{{end}}{{TITLE}}",
			expected:    "{{if .Comment}}{{.Comment}}{{end}}{{.Article.Title}}",
			expectError: false,
		},

		// 異常系テスト
		{
			name:        "小文字の別名記法",
			input:       "{{comment}}",
			expected:    "",
			expectError: true,
			errorMsg:    "別名記法では大文字のみが許可されています。'{{comment}}' の代わりに '{{COMMENT}}' を使用してください",
		},
		{
			name:        "存在しないパラメータ",
			input:       "{{AUTHOR}}",
			expected:    "",
			expectError: true,
			errorMsg:    "存在しないパラメータです: '{{AUTHOR}}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.Convert(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				if aliasErr, ok := err.(*TemplateAliasError); ok {
					assert.Contains(t, aliasErr.Message, tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsUpperCaseOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "大文字のみ",
			input:    "TITLE",
			expected: true,
		},
		{
			name:     "アンダースコア付き",
			input:    "FIXED_MESSAGE",
			expected: true,
		},
		{
			name:     "小文字混在",
			input:    "Title",
			expected: false,
		},
		{
			name:     "全て小文字",
			input:    "title",
			expected: false,
		},
		{
			name:     "数字を含む",
			input:    "TITLE123",
			expected: false,
		},
		{
			name:     "空文字列",
			input:    "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUpperCaseOnly(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTemplateControl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "if文",
			input:    "if .Comment",
			expected: true,
		},
		{
			name:     "else文",
			input:    "else",
			expected: true,
		},
		{
			name:     "end文",
			input:    "end",
			expected: true,
		},
		{
			name:     "range文",
			input:    "range .Items",
			expected: true,
		},
		{
			name:     "非制御構文",
			input:    "TITLE",
			expected: false,
		},
		{
			name:     "大文字のIF",
			input:    "IF .Comment",
			expected: true, // 大文字小文字を区別しない
		},
		{
			name:     "空文字列",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTemplateControl(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateAliasError(t *testing.T) {
	err := &TemplateAliasError{
		InvalidAlias: "{{title}}",
		Message:      "別名記法では大文字のみが許可されています",
	}

	assert.Equal(t, "別名記法では大文字のみが許可されています", err.Error())
	assert.Equal(t, "{{title}}", err.InvalidAlias)
}

func TestGetValidAliases(t *testing.T) {
	t.Run("PromptConverter", func(t *testing.T) {
		converter := NewPromptTemplateAliasConverter()
		aliases := converter.getValidAliases()

		assert.Equal(t, 3, len(aliases))
		// マップの順序は保証されないので、要素の存在だけ確認
		aliasesStr := strings.Join(aliases, " ")
		assert.Contains(t, aliasesStr, "{{TITLE}}")
		assert.Contains(t, aliasesStr, "{{URL}}")
		assert.Contains(t, aliasesStr, "{{CONTENT}}")
	})

	t.Run("SlackConverter", func(t *testing.T) {
		converter := NewSlackTemplateAliasConverter()
		aliases := converter.getValidAliases()

		assert.Equal(t, 5, len(aliases))
		aliasesStr := strings.Join(aliases, " ")
		assert.Contains(t, aliasesStr, "{{TITLE}}")
		assert.Contains(t, aliasesStr, "{{URL}}")
		assert.Contains(t, aliasesStr, "{{CONTENT}}")
		assert.Contains(t, aliasesStr, "{{COMMENT}}")
		assert.Contains(t, aliasesStr, "{{FIXED_MESSAGE}}")
	})
}

