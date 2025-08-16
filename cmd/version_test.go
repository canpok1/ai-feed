package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// TestVersionCommand はversionコマンドの動作を確認する
func TestVersionCommand(t *testing.T) {
	// 出力をキャプチャするバッファを作成
	var buf bytes.Buffer
	cmd := makeVersionCmd()
	cmd.SetOut(&buf)

	// コマンドを実行
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("versionコマンドの実行に失敗: %v", err)
	}

	// 出力の検証
	output := buf.String()
	if !strings.Contains(output, "dev") {
		t.Errorf("期待される出力 'dev' が含まれていません。実際の出力: %s", output)
	}
}

// TestVersionCommandWithCustomVersion はカスタムバージョンでの動作を確認する
func TestVersionCommandWithCustomVersion(t *testing.T) {
	// バージョンを一時的に変更
	originalVersion := version
	version = "v1.2.3"
	defer func() {
		version = originalVersion
	}()

	// 出力をキャプチャするバッファを作成
	var buf bytes.Buffer
	cmd := makeVersionCmd()
	cmd.SetOut(&buf)

	// コマンドを実行
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("versionコマンドの実行に失敗: %v", err)
	}

	// 出力の検証
	output := buf.String()
	if !strings.Contains(output, "v1.2.3") {
		t.Errorf("期待される出力 'v1.2.3' が含まれていません。実際の出力: %s", output)
	}
}
