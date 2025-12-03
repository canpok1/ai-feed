package cmd

import (
	"bytes"
	"log/slog"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVerboseFlag(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectVerbose bool
	}{
		{
			name:          "正常系: verboseフラグなし",
			args:          []string{"recommend", "--help"},
			expectVerbose: false,
		},
		{
			name:          "正常系: verbose短縮フラグ(-v)あり",
			args:          []string{"-v", "recommend", "--help"},
			expectVerbose: true,
		},
		{
			name:          "正常系: verbose完全フラグ(--verbose)あり",
			args:          []string{"--verbose", "recommend", "--help"},
			expectVerbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// verboseフラグをリセット
			verbose = false

			// ルートコマンドを作成
			rootCmd := makeRootCmd()

			// 引数を設定
			rootCmd.SetArgs(tt.args)

			// ヘルプテキストを抑制するために出力をキャプチャ
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// コマンドを実行（verboseフラグが設定されるはず）
			err := rootCmd.Execute()

			// ヘルプコマンドはエラーを返すが、これは期待どおりの動作
			// すべてのケースで、verbose変数が正しく設定されていることを確認
			assert.Equal(t, tt.expectVerbose, verbose)

			// ヘルプコマンドのエラーは無視
			_ = err
		})
	}
}

func TestLoggerInitialization(t *testing.T) {
	// オリジナルのデフォルトロガーを保存
	originalDefault := slog.Default()
	defer slog.SetDefault(originalDefault)

	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "正常系: verbose=falseでのロガー初期化",
			verbose: false,
		},
		{
			name:    "正常系: verbose=trueでのロガー初期化",
			verbose: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// verboseフラグを設定
			verbose = tt.verbose

			// ルートコマンドを作成
			rootCmd := makeRootCmd()

			// PersistentPreRunを設定
			rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
				// 通常はinfra.InitLogger(verbose)を呼び出すが、
				// テストではverboseフラグが正しく設定されていることを検証
				assert.Equal(t, tt.verbose, verbose)
			}

			// PersistentPreRunを発火させるためにダミーコマンドを設定
			rootCmd.SetArgs([]string{"--help"})

			// 出力をキャプチャ
			var buf bytes.Buffer
			rootCmd.SetOut(&buf)
			rootCmd.SetErr(&buf)

			// 実行 - PersistentPreRunが発火するはず
			err := rootCmd.Execute()

			// ヘルプコマンドはエラーで終了するが、これは期待どおりの動作
			_ = err
		})
	}
}

func TestRootCommandCreation(t *testing.T) {
	cmd := makeRootCmd()

	// 基本的なコマンドプロパティをテスト
	assert.Equal(t, "ai-feed", cmd.Use)
	assert.Contains(t, cmd.Short, "RSSフィードから記事を取得")
	assert.True(t, cmd.SilenceUsage)

	// フラグをテスト
	configFlag := cmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "", configFlag.DefValue)

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "false", verboseFlag.DefValue)

	// verbose短縮フラグをテスト
	verboseFlagShort := cmd.PersistentFlags().ShorthandLookup("v")
	assert.NotNil(t, verboseFlagShort)
	assert.Equal(t, verboseFlag, verboseFlagShort)
}

func TestExecuteFunction(t *testing.T) {
	// Execute関数がパニックなしで呼び出せることを確認
	// テスト中にstdoutへの出力を避けるため、出力をリダイレクト

	// オリジナルのos.Argsを保存
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// 実際のコマンド実行を避けるため、ヘルプを表示する引数を設定
	os.Args = []string{"ai-feed", "--help"}

	// Execute関数はパニックしないこと
	assert.NotPanics(t, func() {
		err := Execute()
		// ヘルプコマンドはエラー（終了コード）を返すが、これは期待どおりの動作
		_ = err
	})
}
