package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestMakeUpdateCmd はupdateコマンドの作成を確認する
func TestMakeUpdateCmd(t *testing.T) {
	cmd := makeUpdateCmd()

	// コマンドの基本プロパティを確認
	if cmd.Use != "update" {
		t.Errorf("期待されるUseは 'update' ですが、実際は %q でした", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Shortが空です")
	}

	if cmd.Long == "" {
		t.Error("Longが空です")
	}

	// --checkフラグの存在確認
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Error("--checkフラグが定義されていません")
	}

	if checkFlag.Shorthand != "c" {
		t.Errorf("--checkフラグのショートハンドは 'c' であるべきですが、実際は %q でした", checkFlag.Shorthand)
	}
}

// TestUpdateCommandHelp はupdateコマンドのヘルプ表示を確認する
func TestUpdateCommandHelp(t *testing.T) {
	var buf bytes.Buffer
	cmd := makeUpdateCmd()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("updateコマンドのヘルプ表示に失敗: %v", err)
	}

	output := buf.String()

	// ヘルプ出力に期待する内容が含まれているか確認
	expectedContents := []string{
		"update",
		"ai-feedを最新バージョンに更新します",
		"--check",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(output, expected) {
			t.Errorf("ヘルプ出力に %q が含まれていません", expected)
		}
	}
}

// TestUpdateCommandCheckFlagExists は--checkフラグの存在を確認する
func TestUpdateCommandCheckFlagExists(t *testing.T) {
	cmd := makeUpdateCmd()

	// フラグの存在確認
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Fatal("--checkフラグが定義されていません")
	}

	// デフォルト値の確認
	defaultValue := checkFlag.DefValue
	if defaultValue != "false" {
		t.Errorf("--checkフラグのデフォルト値は 'false' であるべきですが、実際は %q でした", defaultValue)
	}
}
