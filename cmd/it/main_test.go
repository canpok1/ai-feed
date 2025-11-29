// Package it はcmdパッケージの統合テストを提供する
// 統合テストはバイナリをビルドして実際のコマンドを実行することで動作を検証する
package it

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	setupPackage()
	code := m.Run()
	cleanupPackage()
	os.Exit(code)
}
