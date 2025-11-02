package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationBuilder(t *testing.T) {
	t.Run("初期状態が正しい", func(t *testing.T) {
		builder := NewValidationBuilder()

		assert.NotNil(t, builder)
		assert.NotNil(t, builder.errors)
		assert.NotNil(t, builder.warnings)
		assert.Empty(t, builder.errors)
		assert.Empty(t, builder.warnings)
	})
}

func TestValidationBuilder_AddError(t *testing.T) {
	t.Run("エラーメッセージが正しく追加される", func(t *testing.T) {
		builder := NewValidationBuilder()

		builder.AddError("エラー1")
		assert.Len(t, builder.errors, 1)
		assert.Equal(t, "エラー1", builder.errors[0])

		builder.AddError("エラー2")
		assert.Len(t, builder.errors, 2)
		assert.Equal(t, "エラー2", builder.errors[1])
	})

	t.Run("メソッドチェーンが機能する", func(t *testing.T) {
		builder := NewValidationBuilder()

		result := builder.AddError("エラー1").AddError("エラー2")

		assert.Same(t, builder, result)
		assert.Len(t, builder.errors, 2)
	})
}

func TestValidationBuilder_AddWarning(t *testing.T) {
	t.Run("警告メッセージが正しく追加される", func(t *testing.T) {
		builder := NewValidationBuilder()

		builder.AddWarning("警告1")
		assert.Len(t, builder.warnings, 1)
		assert.Equal(t, "警告1", builder.warnings[0])

		builder.AddWarning("警告2")
		assert.Len(t, builder.warnings, 2)
		assert.Equal(t, "警告2", builder.warnings[1])
	})

	t.Run("メソッドチェーンが機能する", func(t *testing.T) {
		builder := NewValidationBuilder()

		result := builder.AddWarning("警告1").AddWarning("警告2")

		assert.Same(t, builder, result)
		assert.Len(t, builder.warnings, 2)
	})
}

func TestValidationBuilder_MergeResult(t *testing.T) {
	t.Run("有効なValidationResultのエラーと警告がマージされる", func(t *testing.T) {
		builder := NewValidationBuilder()
		builder.AddError("既存エラー")
		builder.AddWarning("既存警告")

		result := &ValidationResult{
			IsValid:  true,
			Errors:   []string{},
			Warnings: []string{"新規警告"},
		}

		builder.MergeResult(result)

		assert.Len(t, builder.errors, 1)
		assert.Len(t, builder.warnings, 2)
		assert.Equal(t, "新規警告", builder.warnings[1])
	})

	t.Run("無効なValidationResultのエラーがマージされる", func(t *testing.T) {
		builder := NewValidationBuilder()

		result := &ValidationResult{
			IsValid:  false,
			Errors:   []string{"新規エラー1", "新規エラー2"},
			Warnings: []string{"新規警告"},
		}

		builder.MergeResult(result)

		assert.Len(t, builder.errors, 2)
		assert.Equal(t, "新規エラー1", builder.errors[0])
		assert.Equal(t, "新規エラー2", builder.errors[1])
		assert.Len(t, builder.warnings, 1)
		assert.Equal(t, "新規警告", builder.warnings[0])
	})

	t.Run("nilのValidationResultを処理できる", func(t *testing.T) {
		builder := NewValidationBuilder()
		builder.AddError("既存エラー")

		builder.MergeResult(nil)

		assert.Len(t, builder.errors, 1)
		assert.Equal(t, "既存エラー", builder.errors[0])
	})

	t.Run("メソッドチェーンが機能する", func(t *testing.T) {
		builder := NewValidationBuilder()
		result := &ValidationResult{
			IsValid:  false,
			Errors:   []string{"エラー"},
			Warnings: []string{},
		}

		returnedBuilder := builder.MergeResult(result)

		assert.Same(t, builder, returnedBuilder)
	})
}

func TestValidationBuilder_Build(t *testing.T) {
	t.Run("エラーがない場合_IsValid_trueのValidationResultを返す", func(t *testing.T) {
		builder := NewValidationBuilder()
		builder.AddWarning("警告のみ")

		result := builder.Build()

		require.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, "警告のみ", result.Warnings[0])
	})

	t.Run("エラーがある場合_IsValid_falseのValidationResultを返す", func(t *testing.T) {
		builder := NewValidationBuilder()
		builder.AddError("エラー1")
		builder.AddWarning("警告1")

		result := builder.Build()

		require.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
		assert.Equal(t, "エラー1", result.Errors[0])
		assert.Len(t, result.Warnings, 1)
		assert.Equal(t, "警告1", result.Warnings[0])
	})

	t.Run("警告のみの場合_IsValid_trueのValidationResultを返す", func(t *testing.T) {
		builder := NewValidationBuilder()
		builder.AddWarning("警告1")
		builder.AddWarning("警告2")

		result := builder.Build()

		require.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
		assert.Len(t, result.Warnings, 2)
	})
}

func TestValidateRequired(t *testing.T) {
	t.Run("正常系_空でない文字列", func(t *testing.T) {
		err := ValidateRequired("値", "フィールド名")

		assert.NoError(t, err)
	})

	t.Run("異常系_空文字列", func(t *testing.T) {
		err := ValidateRequired("", "フィールド名")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "フィールド名が設定されていません")
	})
}

func TestValidateURL(t *testing.T) {
	t.Run("正常系_正しいURL形式", func(t *testing.T) {
		testCases := []struct {
			name string
			url  string
		}{
			{name: "HTTP URL", url: "http://example.com"},
			{name: "HTTPS URL", url: "https://example.com"},
			{name: "パス付きURL", url: "https://example.com/path/to/page"},
			{name: "クエリパラメータ付きURL", url: "https://example.com?param=value"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateURL(tc.url, "URL")
				assert.NoError(t, err)
			})
		}
	})

	t.Run("異常系_空文字列", func(t *testing.T) {
		err := ValidateURL("", "URL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URLが設定されていません")
	})

	t.Run("異常系_不正なURL形式", func(t *testing.T) {
		testCases := []struct {
			name string
			url  string
		}{
			{name: "スキームなし", url: "example.com"},
			{name: "不正な形式", url: "not a url"},
			{name: "スペース含む", url: "http://example .com"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				err := ValidateURL(tc.url, "URL")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "URLが正しいURL形式ではありません")
			})
		}
	})

	t.Run("異常系_相対URL", func(t *testing.T) {
		err := ValidateURL("/relative/path", "URL")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URLが正しいURL形式ではありません")
	})
}
