package infra

import (
	_ "embed"
)

//go:embed templates/config.yml
var configTemplateData []byte

//go:embed templates/profile.yml
var profileTemplateData []byte

// GetConfigTemplate はconfig.ymlテンプレートの内容を取得する
func GetConfigTemplate() ([]byte, error) {
	return configTemplateData, nil
}

// GetProfileTemplate はprofile.ymlテンプレートの内容を取得する
func GetProfileTemplate() ([]byte, error) {
	return profileTemplateData, nil
}
