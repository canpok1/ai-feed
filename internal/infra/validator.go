package infra

import (
	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ConfigValidator は設定のバリデーションを行う
type ConfigValidator struct {
	config  *Config
	profile *entity.Profile
}

// NewConfigValidator はConfigValidatorを生成する
func NewConfigValidator(config *Config, profile *entity.Profile) *ConfigValidator {
	return &ConfigValidator{
		config:  config,
		profile: profile,
	}
}

// Validate は設定をバリデーションする
func (v *ConfigValidator) Validate() (*domain.ValidationResult, error) {
	result := &domain.ValidationResult{
		Valid:  true,
		Errors: []domain.ValidationError{},
		Summary: domain.ConfigSummary{
			GeminiConfigured:        false,
			GeminiModel:             "",
			CommentPromptConfigured: false,
			SlackConfigured:         false,
			MisskeyConfigured:       false,
		},
	}

	// AI設定のバリデーション
	v.validateAI(result)

	// プロンプト設定のバリデーション
	v.validatePrompt(result)

	// 出力先設定のバリデーション（設定されている場合のみ）
	v.validateOutput(result)

	// キャッシュ設定のバリデーション（設定されている場合のみ）
	v.validateCache(result)

	// エラーがある場合はValidをfalseに設定
	if len(result.Errors) > 0 {
		result.Valid = false
	}

	return result, nil
}

// validateAI はAI設定をバリデーションする
func (v *ConfigValidator) validateAI(result *domain.ValidationResult) {
	if v.profile.AI == nil {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "ai",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "AI設定が設定されていません",
		})
		return
	}

	if v.profile.AI.Gemini == nil {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "ai.gemini",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Gemini設定が設定されていません",
		})
		return
	}

	gemini := v.profile.AI.Gemini

	// Type のバリデーション
	if gemini.Type == "" {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "ai.gemini.type",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Geminiモデルタイプが設定されていません",
		})
	}

	// APIKey のバリデーション
	if gemini.APIKey.IsEmpty() {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "ai.gemini.api_key",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Gemini APIキーが設定されていません",
		})
	} else if isDummyValue(gemini.APIKey.Value()) {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "ai.gemini.api_key",
			Type:    domain.ValidationErrorTypeDummyValue,
			Message: "Gemini APIキーがダミー値です: \"" + gemini.APIKey.Value() + "\"",
		})
	}

	// サマリーの更新
	if gemini.Type != "" && !gemini.APIKey.IsEmpty() && !isDummyValue(gemini.APIKey.Value()) {
		result.Summary.GeminiConfigured = true
		result.Summary.GeminiModel = gemini.Type
	}
}

// validatePrompt はプロンプト設定をバリデーションする
func (v *ConfigValidator) validatePrompt(result *domain.ValidationResult) {
	if v.profile.Prompt == nil {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "prompt",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "プロンプト設定が設定されていません",
		})
		return
	}

	prompt := v.profile.Prompt

	// CommentPromptTemplate のバリデーション
	if prompt.CommentPromptTemplate == "" {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "comment_prompt_template",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "コメントプロンプトテンプレートが設定されていません",
		})
	}

	// サマリーの更新
	if prompt.CommentPromptTemplate != "" {
		result.Summary.CommentPromptConfigured = true
	}
}

// validateOutput は出力先設定をバリデーションする
func (v *ConfigValidator) validateOutput(result *domain.ValidationResult) {
	if v.profile.Output == nil {
		return
	}

	output := v.profile.Output

	// Slack API設定のバリデーション
	if output.SlackAPI != nil && output.SlackAPI.Enabled {
		v.validateSlackAPI(output.SlackAPI, result)
	}

	// Misskey設定のバリデーション
	if output.Misskey != nil && output.Misskey.Enabled {
		v.validateMisskey(output.Misskey, result)
	}
}

// validateCache はキャッシュ設定をバリデーションする
func (v *ConfigValidator) validateCache(result *domain.ValidationResult) {
	if v.config.Cache == nil {
		return
	}

	// entity化してバリデーション
	cacheEntity, err := v.config.Cache.ToEntity()
	if err != nil {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "cache",
			Type:    domain.ValidationErrorTypeRequired,
			Message: err.Error(),
		})
		return
	}

	if cacheEntity == nil {
		return
	}

	// entity層のバリデーションを実行
	validationResult := cacheEntity.Validate()
	if !validationResult.IsValid {
		for _, errMsg := range validationResult.Errors {
			result.Errors = append(result.Errors, domain.ValidationError{
				Field:   "cache.file_path",
				Type:    domain.ValidationErrorTypeRequired,
				Message: errMsg,
			})
		}
	}
}

// validateSlackAPI はSlack API設定をバリデーションする
func (v *ConfigValidator) validateSlackAPI(slack *entity.SlackAPIConfig, result *domain.ValidationResult) {
	if slack.APIToken.IsEmpty() {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.slack_api.api_token",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Slack APIトークンが設定されていません",
		})
	} else if isDummyValue(slack.APIToken.Value()) {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.slack_api.api_token",
			Type:    domain.ValidationErrorTypeDummyValue,
			Message: "Slack APIトークンがダミー値です: \"" + slack.APIToken.Value() + "\"",
		})
	}

	if slack.Channel == "" {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.slack_api.channel",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Slackチャンネルが設定されていません",
		})
	}

	// サマリーの更新
	if !slack.APIToken.IsEmpty() && !isDummyValue(slack.APIToken.Value()) && slack.Channel != "" {
		result.Summary.SlackConfigured = true
	}
}

// validateMisskey はMisskey設定をバリデーションする
func (v *ConfigValidator) validateMisskey(misskey *entity.MisskeyConfig, result *domain.ValidationResult) {
	if misskey.APIToken.IsEmpty() {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.misskey.api_token",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Misskey APIトークンが設定されていません",
		})
	} else if isDummyValue(misskey.APIToken.Value()) {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.misskey.api_token",
			Type:    domain.ValidationErrorTypeDummyValue,
			Message: "Misskey APIトークンがダミー値です: \"" + misskey.APIToken.Value() + "\"",
		})
	}

	if misskey.APIURL == "" {
		result.Errors = append(result.Errors, domain.ValidationError{
			Field:   "output.misskey.api_url",
			Type:    domain.ValidationErrorTypeRequired,
			Message: "Misskey API URLが設定されていません",
		})
	}

	// サマリーの更新
	if !misskey.APIToken.IsEmpty() && !isDummyValue(misskey.APIToken.Value()) && misskey.APIURL != "" {
		result.Summary.MisskeyConfigured = true
	}
}

// dummyValues はダミー値として認識する文字列のセット
var dummyValues = map[string]struct{}{
	"xxxxxx":                             {},
	"YOUR_MISSKEY_PUBLIC_API_TOKEN_HERE": {},
}

// isDummyValue はダミー値かどうかを判定する
func isDummyValue(value string) bool {
	_, exists := dummyValues[value]
	return exists
}
