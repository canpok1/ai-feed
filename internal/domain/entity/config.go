package entity

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"text/template"
)

// テンプレートキャッシュ（スレッドセーフ）
var templateCache sync.Map

// デフォルト値の定数

type AIConfig struct {
	Gemini *GeminiConfig
}

// Validate はAIConfigの内容をバリデーションする
func (a *AIConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Gemini: 必須項目（nilでない）
	if a.Gemini == nil {
		builder.AddError("Gemini設定が設定されていません")
	} else {
		// GeminiConfig.Validate()を呼び出して、結果を集約
		builder.MergeResult(a.Gemini.Validate())
	}

	return builder.Build()
}

// Merge は他のAIConfigの非nil フィールドで現在のAIConfigをマージする
func (a *AIConfig) Merge(other *AIConfig) {
	if other == nil {
		return
	}
	mergePtr(&a.Gemini, other.Gemini)
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (a AIConfig) LogValue() slog.Value {
	if a.Gemini != nil {
		return slog.GroupValue(
			slog.Any("Gemini", *a.Gemini), // GeminiConfig.LogValue() が呼ばれる
		)
	}
	return slog.GroupValue()
}

type GeminiConfig struct {
	Type   string
	APIKey SecretString
}

// Validate はGeminiConfigの内容をバリデーションする
func (g *GeminiConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// Type: 必須項目（空文字列でない）
	if err := ValidateRequired(g.Type, "Gemini設定のType"); err != nil {
		builder.AddError(err.Error())
	}

	// APIKey: 必須項目（空でない）
	if g.APIKey.IsEmpty() {
		builder.AddError("Gemini APIキーが設定されていません")
	}

	return builder.Build()
}

// Merge は他のGeminiConfigの非空フィールドで現在のGeminiConfigをマージする
func (g *GeminiConfig) Merge(other *GeminiConfig) {
	if other == nil {
		return
	}
	mergeString(&g.Type, other.Type)
	if !other.APIKey.IsEmpty() {
		g.APIKey = other.APIKey
	}
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (g GeminiConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Type", g.Type),
		slog.Any("APIKey", g.APIKey),
	)
}

type PromptConfig struct {
	SystemPrompt           string
	CommentPromptTemplate  string
	SelectorPromptTemplate string
	FixedMessage           string
}

// Validate はPromptConfigの内容をバリデーションする
func (p *PromptConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// SystemPrompt: 必須項目（空文字列でない）
	if err := ValidateRequired(p.SystemPrompt, "システムプロンプト"); err != nil {
		builder.AddError(err.Error())
	}

	// CommentPromptTemplate: 必須項目（空文字列でない）
	if err := ValidateRequired(p.CommentPromptTemplate, "コメントプロンプトテンプレート"); err != nil {
		builder.AddError(err.Error())
	}

	// FixedMessage: 任意項目（空文字列でも可）

	return builder.Build()
}

// BuildCommentPrompt はtext/templateを使用してコメントプロンプトを生成する
func (c *PromptConfig) BuildCommentPrompt(article *Article) (string, error) {
	// 後方互換性のため、古い形式のプレースホルダーを新形式に変換
	templateStr := c.CommentPromptTemplate
	templateStr = strings.ReplaceAll(templateStr, "{{title}}", "{{.Title}}")
	templateStr = strings.ReplaceAll(templateStr, "{{url}}", "{{.Link}}")
	templateStr = strings.ReplaceAll(templateStr, "{{content}}", "{{.Content}}")

	// 別名記法（{{TITLE}}など）を既存記法に変換
	converter := NewPromptTemplateAliasConverter()
	convertedTemplate, err := converter.Convert(templateStr)
	if err != nil {
		// 別名変換でエラーが発生した場合は、エラーを伝播する
		return "", fmt.Errorf("テンプレート変換エラー: %w", err)
	}
	templateStr = convertedTemplate

	// キャッシュからテンプレートを取得
	var tmpl *template.Template
	if cached, ok := templateCache.Load(templateStr); ok {
		tmpl = cached.(*template.Template)
	} else {
		// キャッシュにない場合はパースして保存
		var err error
		tmpl, err = template.New("comment").Parse(templateStr)
		if err != nil {
			// テンプレートの解析に失敗した場合は、エラーを返す
			return "", fmt.Errorf("テンプレート解析エラー: %w", err)
		}
		// パース成功したテンプレートをキャッシュに保存
		templateCache.Store(templateStr, tmpl)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, article)
	if err != nil {
		// テンプレートの実行に失敗した場合も、エラーを返す
		return "", fmt.Errorf("テンプレート実行エラー: %w", err)
	}

	return buf.String(), nil
}

// Merge は他のPromptConfigの非空フィールドで現在のPromptConfigをマージする
func (p *PromptConfig) Merge(other *PromptConfig) {
	if other == nil {
		return
	}
	mergeString(&p.SystemPrompt, other.SystemPrompt)
	mergeString(&p.CommentPromptTemplate, other.CommentPromptTemplate)
	mergeString(&p.SelectorPromptTemplate, other.SelectorPromptTemplate)
	mergeString(&p.FixedMessage, other.FixedMessage)
}

// LogValue はslog出力時に設定値を読みやすく表示するためのメソッド
func (p PromptConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("SystemPromptLength", len(p.SystemPrompt)),
		slog.Int("CommentPromptTemplateLength", len(p.CommentPromptTemplate)),
		slog.Int("SelectorPromptTemplateLength", len(p.SelectorPromptTemplate)),
		slog.String("FixedMessage", p.FixedMessage),
	)
}

type MisskeyConfig struct {
	Enabled         bool
	APIToken        SecretString
	APIURL          string
	MessageTemplate *string
}

// Validate はMisskeyConfigの内容をバリデーションする
func (m *MisskeyConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// APIToken: 必須項目（空でない）
	if m.APIToken.IsEmpty() {
		builder.AddError("Misskey APIトークンが設定されていません")
	}

	// APIURL: 必須項目（空文字列でない）、URL形式であること
	if err := ValidateURL(m.APIURL, "Misskey API URL"); err != nil {
		builder.AddError(err.Error())
	}

	// MessageTemplate: 必須項目
	if m.MessageTemplate == nil || strings.TrimSpace(*m.MessageTemplate) == "" {
		builder.AddError("Misskeyメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nmisskey:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}")
	} else {
		if err := m.validateMisskeyMessageTemplate(*m.MessageTemplate); err != nil {
			builder.AddError(fmt.Sprintf("Misskeyメッセージテンプレートが無効です: %v", err))
		}
	}

	return builder.Build()
}

// validateMisskeyMessageTemplate はMisskeyメッセージテンプレートの構文を検証する
func (m *MisskeyConfig) validateMisskeyMessageTemplate(templateStr string) error {
	// text/templateでパースして構文チェック
	_, err := template.New("misskey_message").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("テンプレート構文エラー: %w", err)
	}

	return nil
}

// Merge は他のMisskeyConfigの非空フィールドで現在のMisskeyConfigをマージする
func (m *MisskeyConfig) Merge(other *MisskeyConfig) {
	if other == nil {
		return
	}
	// bool フィールドはゼロ値チェックが困難なため、常に上書き
	m.Enabled = other.Enabled
	if !other.APIToken.IsEmpty() {
		m.APIToken = other.APIToken
	}
	mergeString(&m.APIURL, other.APIURL)
	if other.MessageTemplate != nil {
		m.MessageTemplate = other.MessageTemplate
	}
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (m MisskeyConfig) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Bool("Enabled", m.Enabled),
		slog.Any("APIToken", m.APIToken),
		slog.String("APIURL", m.APIURL),
	}
	if m.MessageTemplate != nil {
		attrs = append(attrs, slog.Int("MessageTemplateLength", len(*m.MessageTemplate)))
	}
	return slog.GroupValue(attrs...)
}

type SlackAPIConfig struct {
	Enabled         bool
	APIToken        SecretString
	Channel         string
	MessageTemplate *string
	Username        *string
	IconURL         *string
	IconEmoji       *string
}

// Validate はSlackAPIConfigの内容をバリデーションする
func (s *SlackAPIConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// APIToken: 必須項目（空でない）
	if s.APIToken.IsEmpty() {
		builder.AddError("Slack APIトークンが設定されていません")
	}

	// Channel: 必須項目（空文字列でない）
	if err := ValidateRequired(s.Channel, "Slackチャンネル"); err != nil {
		builder.AddError(err.Error())
	}

	// MessageTemplate: 必須項目
	if s.MessageTemplate == nil || strings.TrimSpace(*s.MessageTemplate) == "" {
		builder.AddError("Slackメッセージテンプレートが設定されていません。config.yml または profile.yml で message_template を設定してください。\n設定例:\nslack_api:\n  message_template: |\n    {{if .Comment}}{{.Comment}}\n    {{end}}{{.Article.Title}}\n    {{.Article.Link}}")
	} else {
		if err := s.validateSlackMessageTemplate(*s.MessageTemplate); err != nil {
			builder.AddError(fmt.Sprintf("Slackメッセージテンプレートが無効です: %v", err))
		}
	}

	// IconURL and IconEmoji cannot be set at the same time
	if s.IconURL != nil && *s.IconURL != "" && s.IconEmoji != nil && *s.IconEmoji != "" {
		builder.AddError("Slack設定エラー: icon_urlとicon_emojiを同時に指定することはできません。")
	}

	return builder.Build()
}

// validateSlackMessageTemplate はSlackメッセージテンプレートの構文を検証する
func (s *SlackAPIConfig) validateSlackMessageTemplate(templateStr string) error {
	// text/templateでパースして構文チェック
	_, err := template.New("slack_message").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("テンプレート構文エラー: %w", err)
	}

	return nil
}

// Merge は他のSlackAPIConfigの非空フィールドで現在のSlackAPIConfigをマージする
func (s *SlackAPIConfig) Merge(other *SlackAPIConfig) {
	if other == nil {
		return
	}
	// bool フィールドはゼロ値チェックが困難なため、常に上書き
	s.Enabled = other.Enabled
	if !other.APIToken.IsEmpty() {
		s.APIToken = other.APIToken
	}
	mergeString(&s.Channel, other.Channel)
	if other.MessageTemplate != nil {
		s.MessageTemplate = other.MessageTemplate
	}
	if other.Username != nil {
		s.Username = other.Username
	}
	if other.IconURL != nil {
		s.IconURL = other.IconURL
	}
	if other.IconEmoji != nil {
		s.IconEmoji = other.IconEmoji
	}
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (s SlackAPIConfig) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Bool("Enabled", s.Enabled),
		slog.Any("APIToken", s.APIToken),
		slog.String("Channel", s.Channel),
	}
	if s.MessageTemplate != nil {
		attrs = append(attrs, slog.Int("MessageTemplateLength", len(*s.MessageTemplate)))
	}
	if s.Username != nil {
		attrs = append(attrs, slog.String("Username", *s.Username))
	}
	if s.IconURL != nil {
		attrs = append(attrs, slog.String("IconURL", *s.IconURL))
	}
	if s.IconEmoji != nil {
		attrs = append(attrs, slog.String("IconEmoji", *s.IconEmoji))
	}
	return slog.GroupValue(attrs...)
}

type CacheConfig struct {
	Enabled       bool
	FilePath      string
	MaxEntries    int
	RetentionDays int
}

// Validate はCacheConfigの内容をバリデーションする
func (c *CacheConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// FilePath: 必須項目（空文字列でない）
	if err := ValidateRequired(c.FilePath, "キャッシュファイルパス"); err != nil {
		builder.AddError(err.Error())
	}

	// MaxEntries と RetentionDays は infra.CacheConfig.ToEntity() で
	// デフォルト値が設定されるため、ここでの正数チェックは不要

	return builder.Build()
}

// Merge は他のCacheConfigの非ゼロ値フィールドで現在のCacheConfigをマージする
func (c *CacheConfig) Merge(other *CacheConfig) {
	if other == nil {
		return
	}
	// bool フィールドはゼロ値チェックが困難なため、常に上書き
	c.Enabled = other.Enabled
	mergeString(&c.FilePath, other.FilePath)
	if other.MaxEntries > 0 {
		c.MaxEntries = other.MaxEntries
	}
	if other.RetentionDays > 0 {
		c.RetentionDays = other.RetentionDays
	}
}

// LogValue はslog出力時に設定値を読みやすく表示するためのメソッド
func (c CacheConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("Enabled", c.Enabled),
		slog.String("FilePath", c.FilePath),
		slog.Int("MaxEntries", c.MaxEntries),
		slog.Int("RetentionDays", c.RetentionDays),
	)
}

type Profile struct {
	AI     *AIConfig
	Prompt *PromptConfig
	Output *OutputConfig
}

// Validate はProfileの内容をバリデーションする
func (p *Profile) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// AI: 必須項目（nilでない）
	if p.AI == nil {
		builder.AddError("AI設定が設定されていません")
	} else {
		builder.MergeResult(p.AI.Validate())
	}

	// Prompt: 必須項目（nilでない）
	if p.Prompt == nil {
		builder.AddError("プロンプト設定が設定されていません")
	} else {
		builder.MergeResult(p.Prompt.Validate())
	}

	// Output: 必須項目（nilでない）
	if p.Output == nil {
		builder.AddError("出力設定が設定されていません")
	} else {
		builder.MergeResult(p.Output.Validate())
	}

	return builder.Build()
}

// Merge は他のProfileの非nil フィールドで現在のProfileをマージする
func (p *Profile) Merge(other *Profile) {
	if other == nil {
		return
	}
	mergePtr(&p.AI, other.AI)
	mergePtr(&p.Prompt, other.Prompt)
	mergePtr(&p.Output, other.Output)
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (p Profile) LogValue() slog.Value {
	attrs := []slog.Attr{}
	if p.AI != nil {
		attrs = append(attrs, slog.Any("AI", *p.AI)) // AIConfig.LogValue() が呼ばれる
	}
	if p.Prompt != nil {
		attrs = append(attrs, slog.Any("Prompt", *p.Prompt))
	}
	if p.Output != nil {
		attrs = append(attrs, slog.Any("Output", *p.Output)) // OutputConfig.LogValue() が呼ばれる
	}
	return slog.GroupValue(attrs...)
}

type OutputConfig struct {
	SlackAPI *SlackAPIConfig
	Misskey  *MisskeyConfig
}

// merger はMergeメソッドを持つ型の制約
type merger[T any] interface {
	Merge(T)
}

// mergePtr はポインタフィールドのマージを行うヘルパー関数
func mergePtr[T any, P interface {
	*T
	merger[P]
}](target *P, source P) {
	if source != nil {
		if *target == nil {
			*target = new(T)
		}
		(*target).Merge(source)
	}
}

// mergeString は文字列フィールドのマージを行うヘルパー関数
func mergeString(target *string, source string) {
	if source != "" {
		*target = source
	}
}

// Validate はOutputConfigの内容をバリデーションする
func (o *OutputConfig) Validate() *ValidationResult {
	builder := NewValidationBuilder()

	// SlackAPIとMisskeyの少なくとも一方は設定されている必要がある
	if o.SlackAPI == nil && o.Misskey == nil {
		builder.AddError("SlackAPI設定またはMisskey設定の少なくとも一方が必要です")
	}

	// 設定されているConfigオブジェクトに対してそれぞれのValidate()メソッドを呼び出す
	if o.SlackAPI != nil {
		builder.MergeResult(o.SlackAPI.Validate())
	}

	if o.Misskey != nil {
		builder.MergeResult(o.Misskey.Validate())
	}

	return builder.Build()
}

// Merge は他のOutputConfigの非nil フィールドで現在のOutputConfigをマージする
func (o *OutputConfig) Merge(other *OutputConfig) {
	if other == nil {
		return
	}
	mergePtr(&o.SlackAPI, other.SlackAPI)
	mergePtr(&o.Misskey, other.Misskey)
}

// LogValue はslog出力時に機密情報をマスクするためのメソッド
func (o OutputConfig) LogValue() slog.Value {
	attrs := []slog.Attr{}
	if o.SlackAPI != nil {
		attrs = append(attrs, slog.Any("SlackAPI", *o.SlackAPI)) // SlackAPIConfig.LogValue() が呼ばれる
	}
	if o.Misskey != nil {
		attrs = append(attrs, slog.Any("Misskey", *o.Misskey)) // MisskeyConfig.LogValue() が呼ばれる
	}
	return slog.GroupValue(attrs...)
}

// ValidationResult はバリデーション結果を表現する
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}
