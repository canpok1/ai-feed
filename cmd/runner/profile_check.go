package runner

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
)

// ProfileCheckResult はプロファイル検証の結果を表す構造体
type ProfileCheckResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

// ProfileCheckRunner はprofile checkコマンドのビジネスロジックを実行する構造体
type ProfileCheckRunner struct {
	configPath    string
	stderr        io.Writer
	profileRepoFn func(string) domain.ProfileRepository
}

// NewProfileCheckRunner はProfileCheckRunnerの新しいインスタンスを作成する
func NewProfileCheckRunner(configPath string, stderr io.Writer, profileRepoFn func(string) domain.ProfileRepository) *ProfileCheckRunner {
	return &ProfileCheckRunner{
		configPath:    configPath,
		stderr:        stderr,
		profileRepoFn: profileRepoFn,
	}
}

// Run はprofile checkコマンドのビジネスロジックを実行する
// profilePathが空の場合はconfig.ymlのデフォルトプロファイルのみを検証
// profilePathが指定されている場合は、指定されたプロファイルをconfig.ymlとマージして検証
func (r *ProfileCheckRunner) Run(profilePath string) (*ProfileCheckResult, error) {
	// config.ymlの読み込み
	var config *infra.Config
	var currentProfile *entity.Profile

	configRepo := infra.NewYamlConfigRepository(r.configPath)
	loadedConfig, err := configRepo.Load()
	if err != nil {
		// ファイルが存在しない場合は警告を表示しない
		if _, statErr := os.Stat(r.configPath); !os.IsNotExist(statErr) {
			// ファイルが存在するが読み込み・パースに失敗した場合は警告を表示
			fmt.Fprintf(r.stderr, "警告: %s の読み込みまたは解析に失敗しました。空のデフォルトプロファイルで継続します。\n", r.configPath)
			slog.Warn("Failed to load or parse config file, continuing with empty default profile", "config_path", r.configPath, "error", err)
		}
	} else {
		config = loadedConfig
	}

	// デフォルトプロファイルの初期化
	if config != nil && config.DefaultProfile != nil {
		var err error
		currentProfile, err = config.DefaultProfile.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to process default profile: %w", err)
		}
	} else {
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

	// バリデーション実行
	validationResult := currentProfile.Validate()

	// 結果を返す
	return &ProfileCheckResult{
		IsValid:  validationResult.IsValid,
		Errors:   validationResult.Errors,
		Warnings: validationResult.Warnings,
	}, nil
}
