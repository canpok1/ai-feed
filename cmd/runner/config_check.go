package runner

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra" // depcheck:allow TODO(#333): cmd/runner を internal/app に移動後、infra依存を解消する
)

// ConfigCheckParams はconfig checkコマンドの実行パラメータを表す構造体
type ConfigCheckParams struct {
	ProfilePath string
	VerboseFlag bool
}

// ConfigCheckRunner はconfig checkコマンドのビジネスロジックを実行する構造体
type ConfigCheckRunner struct {
	configPath    string
	stdout        io.Writer
	stderr        io.Writer
	profileRepoFn func(string) domain.ProfileRepository
}

// NewConfigCheckRunner はConfigCheckRunnerの新しいインスタンスを作成する
func NewConfigCheckRunner(configPath string, stdout io.Writer, stderr io.Writer, profileRepoFn func(string) domain.ProfileRepository) *ConfigCheckRunner {
	return &ConfigCheckRunner{
		configPath:    configPath,
		stdout:        stdout,
		stderr:        stderr,
		profileRepoFn: profileRepoFn,
	}
}

// Run はconfig checkコマンドのビジネスロジックを実行する
func (r *ConfigCheckRunner) Run(params *ConfigCheckParams) error {
	slog.Debug("Starting config check command")

	// 設定ファイルの読み込み
	slog.Debug("Loading config", "config_path", r.configPath)
	config, configLoadErr := infra.NewYamlConfigRepository(r.configPath).Load()
	if configLoadErr != nil {
		fmt.Fprintf(r.stderr, "エラー: 設定ファイルの読み込みに失敗しました: %s\n", r.configPath)
		fmt.Fprintln(r.stderr, "config.ymlの構文を確認してください。ai-feed init で新しい設定ファイルを生成できます。")
		slog.Error("Failed to load config", "config_path", r.configPath, "error", configLoadErr)
		return fmt.Errorf("failed to load config: %w", configLoadErr)
	}

	// デフォルトプロファイルをentity.Profileに変換
	var currentProfile *entity.Profile
	if config.DefaultProfile == nil {
		currentProfile = &entity.Profile{}
	} else {
		p, err := config.DefaultProfile.ToEntity()
		if err != nil {
			return fmt.Errorf("failed to convert profile to entity: %w", err)
		}
		currentProfile = p
	}

	// プロファイルファイルが指定されている場合は読み込んでマージ
	if params.ProfilePath != "" {
		slog.Debug("Loading profile", "profile_path", params.ProfilePath)
		profileRepo := r.profileRepoFn(params.ProfilePath)
		loadedProfile, loadProfileErr := profileRepo.LoadProfile()
		if loadProfileErr != nil {
			fmt.Fprintf(r.stderr, "エラー: プロファイルファイルの読み込みに失敗しました: %s\n", params.ProfilePath)
			fmt.Fprintln(r.stderr, "プロファイルファイルの形式を確認してください。")
			slog.Error("Failed to load profile", "profile_path", params.ProfilePath, "error", loadProfileErr)
			return fmt.Errorf("failed to load profile from %s: %w", params.ProfilePath, loadProfileErr)
		}
		currentProfile.Merge(loadedProfile)
	}

	// バリデーションを実行
	validator := infra.NewConfigValidator(config, currentProfile)
	result, validateErr := validator.Validate()
	if validateErr != nil {
		return fmt.Errorf("failed to validate config: %w", validateErr)
	}

	// バリデーション結果を出力
	printValidationResult(r.stdout, r.stderr, result, params.VerboseFlag)

	// バリデーション失敗時は終了コード1
	if !result.Valid {
		return fmt.Errorf("設定ファイルのバリデーションに失敗しました")
	}

	slog.Debug("Config check command completed successfully")
	return nil
}

// printValidationResult はバリデーション結果を出力する（統一形式: 1行目=処理完了報告、2行目以降=結果報告）
func printValidationResult(stdout, stderr io.Writer, result *domain.ValidationResult, verboseFlag bool) {
	if result.Valid {
		// 成功時はstdoutに出力
		fmt.Fprintln(stdout, "設定の検証が完了しました")
		fmt.Fprintln(stdout, "問題ありませんでした")
		if verboseFlag {
			printSummary(stdout, result.Summary)
		}
	} else {
		// 失敗時はすべてstderrに出力
		fmt.Fprintln(stderr, "設定の検証が完了しました")
		fmt.Fprintln(stderr, "以下の問題があります：")
		for _, err := range result.Errors {
			fmt.Fprintf(stderr, "- %s: %s\n", err.Field, err.Message)
		}
	}
}

// printSummary は設定のサマリー情報を出力する
func printSummary(stdout io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(stdout, "")
	fmt.Fprintln(stdout, "【設定サマリー】")
	printAISummary(stdout, summary)
	printPromptSummary(stdout, summary)
	printOutputSummary(stdout, summary)
	printCacheSummary(stdout, summary)
}

// printAISummary はAI設定のサマリーを出力する
func printAISummary(stdout io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(stdout, "AI設定:")
	if summary.GeminiConfigured {
		fmt.Fprintf(stdout, "  - Gemini API: 設定済み（モデル: %s）\n", summary.GeminiModel)
	} else {
		fmt.Fprintln(stdout, "  - Gemini API: 未設定")
	}
}

// printPromptSummary はプロンプト設定のサマリーを出力する
func printPromptSummary(stdout io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(stdout, "プロンプト設定:")
	if summary.SystemPromptConfigured {
		fmt.Fprintln(stdout, "  - システムプロンプト: 設定済み")
	} else {
		fmt.Fprintln(stdout, "  - システムプロンプト: 未設定")
	}
	if summary.CommentPromptConfigured {
		fmt.Fprintln(stdout, "  - コメントプロンプト: 設定済み")
	} else {
		fmt.Fprintln(stdout, "  - コメントプロンプト: 未設定")
	}
	if summary.FixedMessageConfigured {
		fmt.Fprintln(stdout, "  - 固定メッセージ: 設定済み")
	} else {
		fmt.Fprintln(stdout, "  - 固定メッセージ: 未設定")
	}
}

// printOutputSummary は出力設定のサマリーを出力する
func printOutputSummary(stdout io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(stdout, "出力設定:")
	if summary.SlackConfigured {
		fmt.Fprintln(stdout, "  - Slack API: 有効")
		fmt.Fprintf(stdout, "    - チャンネル: %s\n", summary.SlackChannel)
		if summary.SlackMessageTemplateConfigured {
			fmt.Fprintln(stdout, "    - メッセージテンプレート: 設定済み")
		} else {
			fmt.Fprintln(stdout, "    - メッセージテンプレート: 未設定")
		}
	} else {
		fmt.Fprintln(stdout, "  - Slack API: 無効")
	}
	if summary.MisskeyConfigured {
		fmt.Fprintln(stdout, "  - Misskey: 有効")
		fmt.Fprintf(stdout, "    - API URL: %s\n", summary.MisskeyAPIURL)
		if summary.MisskeyMessageTemplateConfigured {
			fmt.Fprintln(stdout, "    - メッセージテンプレート: 設定済み")
		} else {
			fmt.Fprintln(stdout, "    - メッセージテンプレート: 未設定")
		}
	} else {
		fmt.Fprintln(stdout, "  - Misskey: 無効")
	}
}

// printCacheSummary はキャッシュ設定のサマリーを出力する
func printCacheSummary(stdout io.Writer, summary domain.ConfigSummary) {
	fmt.Fprintln(stdout, "キャッシュ設定:")
	if summary.CacheEnabled {
		fmt.Fprintln(stdout, "  - キャッシュ: 有効")
		fmt.Fprintf(stdout, "    - ファイルパス: %s\n", summary.CacheFilePath)
		fmt.Fprintf(stdout, "    - 最大エントリ数: %d\n", summary.CacheMaxEntries)
		fmt.Fprintf(stdout, "    - 保持期間: %d日\n", summary.CacheRetentionDays)
	} else {
		fmt.Fprintln(stdout, "  - キャッシュ: 無効")
	}
}
