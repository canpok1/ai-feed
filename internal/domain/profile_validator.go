package domain

import (
	"strings"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ValidationResult はプロファイルバリデーションの結果を表現する
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ProfileValidator はプロファイルファイルのバリデーションを行うインターフェース
type ProfileValidator interface {
	// Validate はプロファイルの内容をバリデーションする
	Validate(profile *entity.Profile) *ValidationResult
}

// ProfileValidatorImpl はProfileValidatorの実装
type ProfileValidatorImpl struct{}

// NewProfileValidator はProfileValidatorImplの新しいインスタンスを作成する
func NewProfileValidator() ProfileValidator {
	return &ProfileValidatorImpl{}
}

// Validate はプロファイルの内容をバリデーションする
func (v *ProfileValidatorImpl) Validate(profile *entity.Profile) *ValidationResult {
	var errors []string
	var warnings []string

	// 必須項目のバリデーション（エラー扱い）
	errors = append(errors, v.validateRequiredFields(profile)...)

	// 警告項目のバリデーション
	warnings = append(warnings, v.validateWarningFields(profile)...)

	return &ValidationResult{
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
		profile.AI.Gemini.APIKey == "YOUR_GEMINI_API_KEY_HERE" {
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

	return errors
}

// validateWarningFields は警告項目のバリデーションを行う
func (v *ProfileValidatorImpl) validateWarningFields(profile *entity.Profile) []string {
	var warnings []string

	// Slack APIトークンの検証
	if profile.Output == nil || profile.Output.SlackAPI == nil ||
		profile.Output.SlackAPI.APIToken == "" ||
		profile.Output.SlackAPI.APIToken == "xoxb-YOUR_SLACK_API_TOKEN_HERE" {
		warnings = append(warnings, "Slack API token is not configured")
	}

	// Slack チャンネルの検証
	if profile.Output == nil || profile.Output.SlackAPI == nil ||
		profile.Output.SlackAPI.Channel == "" {
		warnings = append(warnings, "Slack channel is not configured")
	}

	// Misskey APIトークンの検証
	if profile.Output == nil || profile.Output.Misskey == nil ||
		profile.Output.Misskey.APIToken == "" ||
		profile.Output.Misskey.APIToken == "YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE" {
		warnings = append(warnings, "Misskey API token is not configured")
	}

	// Misskey API URLの検証
	if profile.Output == nil || profile.Output.Misskey == nil ||
		profile.Output.Misskey.APIURL == "" {
		warnings = append(warnings, "Misskey API URL is not configured")
	}

	return warnings
}

// MaskSensitiveData はAPIキーなどの機密情報をマスクする
func MaskSensitiveData(value string) string {
	if value == "" {
		return ""
	}
	
	// デフォルト値の場合はそのまま返す
	defaultValues := []string{
		"YOUR_GEMINI_API_KEY_HERE",
		"xoxb-YOUR_SLACK_API_TOKEN_HERE",
		"YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE",
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