package domain

// ValidationErrorType はバリデーションエラーの種別を表す
type ValidationErrorType int

const (
	// ValidationErrorTypeRequired は必須項目が未設定のエラー
	ValidationErrorTypeRequired ValidationErrorType = iota
	// ValidationErrorTypeDummyValue はダミー値が設定されているエラー
	ValidationErrorTypeDummyValue
	// ValidationErrorTypeInvalid は不正な値が設定されているエラー
	ValidationErrorTypeInvalid
)

// ValidationError はバリデーションエラーの詳細を表す
type ValidationError struct {
	// Field はエラーが発生したフィールド名
	Field string
	// Type はエラーの種別
	Type ValidationErrorType
	// Message は詳細なエラーメッセージ
	Message string
}

// ConfigSummary は設定のサマリー情報を表す（verbose用）
type ConfigSummary struct {
	// GeminiConfigured はGemini APIの設定状態
	GeminiConfigured bool
	// GeminiModel は設定されているGeminiモデル
	GeminiModel string
	// SystemPromptConfigured はシステムプロンプトの設定状態
	SystemPromptConfigured bool
	// CommentPromptConfigured はコメントプロンプトの設定状態
	CommentPromptConfigured bool
	// FixedMessageConfigured は固定メッセージの設定状態
	FixedMessageConfigured bool
	// SlackConfigured はSlack APIの設定状態
	SlackConfigured bool
	// MisskeyConfigured はMisskeyの設定状態
	MisskeyConfigured bool
}

// ValidationResult はバリデーション結果を表す
type ValidationResult struct {
	// Valid はバリデーションが成功したかどうか
	Valid bool
	// Errors はバリデーションエラーのリスト
	Errors []ValidationError
	// Summary は設定のサマリー情報
	Summary ConfigSummary
}

// Validator は設定のバリデーションを担当するインターフェース
type Validator interface {
	Validate() (*ValidationResult, error)
}
