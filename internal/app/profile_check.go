package app

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ProfileCheckResult はプロファイル検証の結果を表す構造体
type ProfileCheckResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

// ProfileCheckRunner はprofile checkコマンドのビジネスロジックを実行する構造体
type ProfileCheckRunner struct {
	configRepo    domain.ConfigRepository
	stderr        io.Writer
	profileRepoFn func(string) domain.ProfileRepository
}

// NewProfileCheckRunner はProfileCheckRunnerの新しいインスタンスを作成する
func NewProfileCheckRunner(configRepo domain.ConfigRepository, stderr io.Writer, profileRepoFn func(string) domain.ProfileRepository) *ProfileCheckRunner {
	return &ProfileCheckRunner{
		configRepo:    configRepo,
		stderr:        stderr,
		profileRepoFn: profileRepoFn,
	}
}

// Run はprofile checkコマンドのビジネスロジックを実行する
// profilePathが空の場合はconfig.ymlのデフォルトプロファイルのみを検証
// profilePathが指定されている場合は、指定されたプロファイルをconfig.ymlとマージして検証
func (r *ProfileCheckRunner) Run(profilePath string) (*ProfileCheckResult, error) {
	// config.ymlの読み込み
	var currentProfile *entity.Profile

	loadedConfig, err := r.configRepo.Load()
	switch {
	case err != nil:
		// ファイルが存在しない場合は警告を表示しない（LoadedConfigがnilを返すため）
		fmt.Fprintf(r.stderr, "警告: 設定ファイルの読み込みまたは解析に失敗しました。空のデフォルトプロファイルで継続します。\n")
		slog.Warn("Failed to load or parse config file, continuing with empty default profile", "error", err)
		currentProfile = &entity.Profile{}
	case loadedConfig.DefaultProfile != nil:
		currentProfile = loadedConfig.DefaultProfile
	default:
		// 存在しない場合は空のプロファイルを使用
		currentProfile = &entity.Profile{}
	}

	// 引数が指定されている場合は指定ファイルとマージ
	if profilePath != "" {
		// ファイルの存在確認
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("プロファイルファイルが見つかりません: %s", profilePath)
		} else if err != nil {
			return nil, fmt.Errorf("ファイルへのアクセスに失敗しました: %w", err)
		}

		// 指定されたプロファイルファイルの読み込み
		profileRepo := r.profileRepoFn(profilePath)
		loadedProfile, err := profileRepo.LoadProfile()
		if err != nil {
			return nil, fmt.Errorf("プロファイルの読み込みに失敗しました: %w", err)
		}

		// デフォルトプロファイルとマージ
		currentProfile.Merge(loadedProfile)
	}

	// 進行状況メッセージ: AI設定確認
	fmt.Fprintln(r.stderr, "AI設定を確認しています...")

	// バリデーション実行
	validationResult := currentProfile.Validate()

	// 結果を返す
	return &ProfileCheckResult{
		IsValid:  validationResult.IsValid,
		Errors:   validationResult.Errors,
		Warnings: validationResult.Warnings,
	}, nil
}
