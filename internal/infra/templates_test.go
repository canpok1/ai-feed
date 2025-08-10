package infra

import (
	"strings"
	"testing"
)

func TestGetConfigTemplate(t *testing.T) {
	// GetConfigTemplate()のテスト
	data, err := GetConfigTemplate()
	if err != nil {
		t.Fatalf("GetConfigTemplate() returned error: %v", err)
	}

	content := string(data)

	// 期待される内容が含まれていることを確認
	expectedStrings := []string{
		"# AI Feedの設定ファイル",
		"default_profile:",
		"ai:",
		"gemini:",
		"system_prompt:",
		"comment_prompt_template:",
		"output:",
		"slack_api:",
		"misskey:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("GetConfigTemplate() result does not contain expected string: %s", expected)
		}
	}
}

func TestGetProfileTemplate(t *testing.T) {
	// GetProfileTemplate()のテスト
	data, err := GetProfileTemplate()
	if err != nil {
		t.Fatalf("GetProfileTemplate() returned error: %v", err)
	}

	content := string(data)

	// 期待される内容が含まれていることを確認
	expectedStrings := []string{
		"# AI Feedのプロファイル設定ファイル",
		"ai:",
		"gemini:",
		"system_prompt:",
		"comment_prompt_template:",
		"output:",
		"slack_api:",
		"misskey:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(content, expected) {
			t.Errorf("GetProfileTemplate() result does not contain expected string: %s", expected)
		}
	}
}
