package profile

import (
	"fmt"
	"os"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
)

// YamlProfileRepository はYAML形式でプロファイルを永続化する実装
type YamlProfileRepository struct {
	filePath string
}

// NewYamlProfileRepository は新しいYamlProfileRepositoryを作成する
func NewYamlProfileRepository(filePath string) domain.ProfileRepository {
	return &YamlProfileRepository{
		filePath: filePath,
	}
}

// NewYamlProfileRepositoryImpl は具体的な実装を返す（infra.Profileを直接扱う必要がある内部パッケージやテストで使用）
func NewYamlProfileRepositoryImpl(filePath string) *YamlProfileRepository {
	return &YamlProfileRepository{
		filePath: filePath,
	}
}

// LoadProfile はプロファイルをファイルから読み込み、entity.Profileを返す（domain.ProfileRepositoryインターフェース用）
func (r *YamlProfileRepository) LoadProfile() (*entity.Profile, error) {
	// YAMLファイルから直接infra.Profileを読み込み
	infraProfile, err := infra.LoadYAML[infra.Profile](r.filePath)
	if err != nil {
		return nil, err
	}
	// entity.Profileに変換
	return infraProfile.ToEntity()
}

// SaveProfileTemplate はテンプレートを使用してコメント付きprofile.ymlファイルを生成する
func (r *YamlProfileRepository) SaveProfileTemplate() error {
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("profile file already exists: %s", r.filePath)
		}
		return fmt.Errorf("failed to create profile file: %s, %w", r.filePath, err)
	}
	defer file.Close()

	// 埋め込まれたYAMLファイルの内容を取得してファイルに書き込み
	templateData, err := infra.GetProfileTemplate()
	if err != nil {
		return fmt.Errorf("failed to get profile template: %w", err)
	}

	_, err = file.Write(templateData)
	if err != nil {
		return fmt.Errorf("failed to write profile template: %w", err)
	}

	return nil
}
