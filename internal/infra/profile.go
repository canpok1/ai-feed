package infra

type YamlProfileRepository struct {
	filePath string
}

func NewYamlProfileRepository(filePath string) *YamlProfileRepository {
	return &YamlProfileRepository{
		filePath: filePath,
	}
}

func (r *YamlProfileRepository) LoadProfile() (*Profile, error) {
	return loadYaml[Profile](r.filePath)
}
