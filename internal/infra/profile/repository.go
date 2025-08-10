package profile

import (
	"fmt"
	"os"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/canpok1/ai-feed/internal/infra"
	"gopkg.in/yaml.v3"
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
	infraProfile, err := r.LoadInfraProfile()
	if err != nil {
		return nil, err
	}
	return infraProfile.ToEntity(), nil
}

// LoadInfraProfile はプロファイルをファイルから読み込み、infra.Profileを返す（内部実装用）
func (r *YamlProfileRepository) LoadInfraProfile() (*infra.Profile, error) {
	return loadYaml[infra.Profile](r.filePath)
}

// SaveProfileWithTemplate はテンプレートを使用してコメント付きprofile.ymlファイルを生成する
func (r *YamlProfileRepository) SaveProfileWithTemplate() error {
	// Use O_WRONLY|O_CREATE|O_EXCL to atomically create the file only if it doesn't exist.
	file, err := os.OpenFile(r.filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("profile file already exists: %s", r.filePath)
		}
		return fmt.Errorf("failed to create profile file: %s, %w", r.filePath, err)
	}
	defer file.Close()

	// テンプレートを実行してファイルに書き込み
	tmpl, err := template.New("profile").Parse(profileYmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse profile template: %w", err)
	}

	err = tmpl.Execute(file, nil)
	if err != nil {
		return fmt.Errorf("failed to execute profile template: %w", err)
	}

	return nil
}

// SaveProfile はプロファイルをファイルに保存する
func (r *YamlProfileRepository) SaveProfile(profile *infra.Profile) error {
	data, err := yaml.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal profile to YAML: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile to file %q: %w", r.filePath, err)
	}
	return nil
}

// loadYaml はYAMLファイルを読み込んで指定された型にデコードする
func loadYaml[T any](filePath string) (*T, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", filePath, err)
	}

	var result T
	if err := yaml.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML from file %q: %w", filePath, err)
	}

	return &result, nil
}
