package runner

import (
	"fmt"
	"io"
	"os"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/canpok1/ai-feed/internal/infra/profile"
)

// ProfileCheckResult はプロファイル検証の結果を表す構造体
type ProfileCheckResult struct {
	IsValid  bool
	Errors   []string
	Warnings []string
}

// ProfileCheckRunner はprofile checkコマンドのビジネスロジックを実行する構造体
type ProfileCheckRunner struct {
	configPath string
	stderr     io.Writer
}

// NewProfileCheckRunner はProfileCheckRunnerの新しいインスタンスを作成する
func NewProfileCheckRunner(configPath string, stderr io.Writer) *ProfileCheckRunner {
	return &ProfileCheckRunner{
		configPath: configPath,
		stderr:     stderr,
	}
}

// Run はprofile checkコマンドのビジネスロジックを実行する
// profilePathが空の場合はconfig.ymlのデフォルトプロファイルのみを検証
// profilePathが指定されている場合は、指定されたプロファイルをconfig.ymlとマージして検証
func (r *ProfileCheckRunner) Run(profilePath string) (*ProfileCheckResult, error) {
	// config.ymlの読み込み
	var config *infra.Config
	var currentProfile infra.Profile

	configRepo := infra.NewYamlConfigRepository(r.configPath)
	loadedConfig, err := configRepo.Load()
	if err != nil {
		// ファイルが存在しない場合は警告を表示しない
		if _, statErr := os.Stat(r.configPath); !os.IsNotExist(statErr) {
			// ファイルが存在するが読み込み・パースに失敗した場合は警告を表示
			fmt.Fprintf(r.stderr, "警告: %s の読み込みまたは解析に失敗しました。空のデフォルトプロファイルで続行します。エラー: %v\n", r.configPath, err)
		}
	} else {
		config = loadedConfig
	}

	// デフォルトプロファイルの取得
	if config != nil && config.DefaultProfile != nil {
		currentProfile = *config.DefaultProfile
	}
	// 存在しない場合は空のプロファイルを使用（ゼロ値）

	// 引数が指定されている場合は指定ファイルとマージ
	if profilePath != "" {
		// ファイルの存在確認
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("プロファイルファイルが見つかりません: %s", profilePath)
		} else if err != nil {
			return nil, fmt.Errorf("ファイルへのアクセスに失敗しました: %w", err)
		}

		// 指定されたプロファイルファイルの読み込み
		loadedInfraProfile, err := profile.NewYamlProfileRepositoryImpl(profilePath).LoadInfraProfile()
		if err != nil {
			return nil, fmt.Errorf("プロファイルの読み込みに失敗しました: %w", err)
		}

		// config.ymlが存在する場合はマージ、存在しない場合はloadedInfraProfileをそのまま使用
		if config != nil && config.DefaultProfile != nil {
			// デフォルトプロファイルとマージ
			currentProfile.Merge(loadedInfraProfile)
		} else {
			// config.ymlが存在しない場合は、読み込んだプロファイルをそのまま使用
			currentProfile = *loadedInfraProfile
		}
	}

	// マージ後のプロファイルをentity.Profileに変換してバリデーション
	entityProfile, err := currentProfile.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to process profile: %w", err)
	}
	validationResult := entityProfile.Validate()

	// 結果を返す
	return &ProfileCheckResult{
		IsValid:  validationResult.IsValid,
		Errors:   validationResult.Errors,
		Warnings: validationResult.Warnings,
	}, nil
}
