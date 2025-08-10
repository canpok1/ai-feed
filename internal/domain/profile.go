package domain

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ProfileRepository はプロファイルの永続化を担当するインターフェース
type ProfileRepository interface {
	LoadProfile() (*entity.Profile, error)
}

// ValidateSlackMessageTemplate はSlackメッセージテンプレートの構文を検証する
func ValidateSlackMessageTemplate(templateStr string) error {
	// 空文字列や空白のみの場合はエラーとしない（デフォルトテンプレートが使用される）
	if strings.TrimSpace(templateStr) == "" {
		return nil
	}

	// text/templateでパースして構文チェック
	_, err := template.New("slack_message").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("テンプレート構文エラー: %w", err)
	}

	return nil
}

// MaskSensitiveData はAPIキーなどの機密情報をマスクする
func MaskSensitiveData(value string) string {
	if value == "" {
		return ""
	}

	// デフォルト値の場合はそのまま返す
	defaultValues := []string{
		entity.DefaultGeminiAPIKey,
		entity.DefaultSlackAPIToken,
		entity.DefaultMisskeyAPIToken,
	}

	for _, defaultVal := range defaultValues {
		if value == defaultVal {
			return value
		}
	}

	// 実際の値の場合はマスクする
	if len(value) <= 8 {
		return strings.Repeat("*", len(value))
	}

	// 最初の4文字と最後の4文字を表示、中間をマスク
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
