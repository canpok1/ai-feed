package entity

import (
	"fmt"
	"net/url"
)

// ValidationBuilder はバリデーション結果を構築するためのヘルパー
type ValidationBuilder struct {
	errors   []string
	warnings []string
}

// NewValidationBuilder は新しいValidationBuilderを作成する
func NewValidationBuilder() *ValidationBuilder {
	return &ValidationBuilder{
		errors:   make([]string, 0),
		warnings: make([]string, 0),
	}
}

// AddError はエラーメッセージを追加する
func (vb *ValidationBuilder) AddError(message string) *ValidationBuilder {
	vb.errors = append(vb.errors, message)
	return vb
}

// AddWarning は警告メッセージを追加する
func (vb *ValidationBuilder) AddWarning(message string) *ValidationBuilder {
	vb.warnings = append(vb.warnings, message)
	return vb
}

// MergeResult は他のValidationResultの結果をマージする
func (vb *ValidationBuilder) MergeResult(result *ValidationResult) *ValidationBuilder {
	if result != nil {
		if !result.IsValid {
			vb.errors = append(vb.errors, result.Errors...)
		}
		vb.warnings = append(vb.warnings, result.Warnings...)
	}
	return vb
}

// Build はValidationResultを構築する
func (vb *ValidationBuilder) Build() *ValidationResult {
	return &ValidationResult{
		IsValid:  len(vb.errors) == 0,
		Errors:   vb.errors,
		Warnings: vb.warnings,
	}
}

// ValidateRequired は必須項目の文字列が空でないことを検証する
func ValidateRequired(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%sが設定されていません", fieldName)
	}
	return nil
}



// ValidateURL はURLが正しい形式であることを検証する
func ValidateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return fmt.Errorf("%sが設定されていません", fieldName)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil || !parsedURL.IsAbs() {
		return fmt.Errorf("%sが正しいURL形式ではありません", fieldName)
	}

	return nil
}
