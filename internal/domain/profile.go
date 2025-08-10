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

// ProfileValidator はプロファイルファイルのバリデーションを行うインターフェース
type ProfileValidator interface {
	// Validate はプロファイルの内容をバリデーションする
	Validate(profile *entity.Profile) *entity.ValidationResult
}

// ProfileValidatorImpl はProfileValidatorの実装
type ProfileValidatorImpl struct{}

// NewProfileValidator はProfileValidatorImplの新しいインスタンスを作成する
func NewProfileValidator() ProfileValidator {
	return &ProfileValidatorImpl{}
}

// Validate はプロファイルの内容をバリデーションする
func (v *ProfileValidatorImpl) Validate(profile *entity.Profile) *entity.ValidationResult {
	var errors []string
	var warnings []string

	// 必須項目のバリデーション（エラー扱い）
	errors = append(errors, v.validateRequiredFields(profile)...)

	// 警告項目のバリデーション
	warnings = append(warnings, v.validateWarningFields(profile)...)

	return &entity.ValidationResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
	}
}

// validateRequiredFields は必須項目のバリデーションを行う
func (v *ProfileValidatorImpl) validateRequiredFields(profile *entity.Profile) []string {
	var errors []string

	// Gemini APIキーの検証
	if profile.AI == nil || profile.AI.Gemini == nil ||
		profile.AI.Gemini.APIKey == "" ||
		profile.AI.Gemini.APIKey == entity.DefaultGeminiAPIKey {
		errors = append(errors, "Gemini API key is not configured")
	}

	// システムプロンプトの検証
	if profile.Prompt == nil || profile.Prompt.SystemPrompt == "" {
		errors = append(errors, "System prompt is not configured")
	}

	// コメントプロンプトテンプレートの検証
	if profile.Prompt == nil || profile.Prompt.CommentPromptTemplate == "" {
		errors = append(errors, "Comment prompt template is not configured")
	}

	// Slackメッセージテンプレートの検証
	if profile.Output != nil && profile.Output.SlackAPI != nil && profile.Output.SlackAPI.MessageTemplate != nil {
		if err := v.validateSlackMessageTemplate(*profile.Output.SlackAPI.MessageTemplate); err != nil {
			errors = append(errors, fmt.Sprintf("Slack message template is invalid: %v", err))
		}
	}

	return errors
}

// validateWarningFields は警告項目のバリデーションを行う
func (v *ProfileValidatorImpl) validateWarningFields(profile *entity.Profile) []string {
	var warnings []string

	// Output設定が存在しない場合は早期リターン
	if profile.Output == nil {
		warnings = append(warnings, "Slack API token is not configured")
		warnings = append(warnings, "Slack channel is not configured")
		warnings = append(warnings, "Misskey API token is not configured")
		warnings = append(warnings, "Misskey API URL is not configured")
		return warnings
	}

	// Slack設定の検証
	if profile.Output.SlackAPI == nil {
		warnings = append(warnings, "Slack API token is not configured")
		warnings = append(warnings, "Slack channel is not configured")
	} else {
		// Slack APIトークンの検証
		if profile.Output.SlackAPI.APIToken == "" ||
			profile.Output.SlackAPI.APIToken == entity.DefaultSlackAPIToken {
			warnings = append(warnings, "Slack API token is not configured")
		}

		// Slack チャンネルの検証
		if profile.Output.SlackAPI.Channel == "" {
			warnings = append(warnings, "Slack channel is not configured")
		}
	}

	// Misskey設定の検証
	if profile.Output.Misskey == nil {
		warnings = append(warnings, "Misskey API token is not configured")
		warnings = append(warnings, "Misskey API URL is not configured")
	} else {
		// Misskey APIトークンの検証
		if profile.Output.Misskey.APIToken == "" ||
			profile.Output.Misskey.APIToken == entity.DefaultMisskeyAPIToken {
			warnings = append(warnings, "Misskey API token is not configured")
		}

		// Misskey API URLの検証
		if profile.Output.Misskey.APIURL == "" {
			warnings = append(warnings, "Misskey API URL is not configured")
		}
	}

	return warnings
}

// validateSlackMessageTemplate はSlackメッセージテンプレートの構文を検証する
func (v *ProfileValidatorImpl) validateSlackMessageTemplate(templateStr string) error {
	return ValidateSlackMessageTemplate(templateStr)
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
