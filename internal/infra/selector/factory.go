package selector

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// ArticleSelectorFactory は ArticleSelector を生成するファクトリ
type ArticleSelectorFactory struct{}

// NewArticleSelectorFactory は新しいファクトリを作成する
func NewArticleSelectorFactory() *ArticleSelectorFactory {
	return &ArticleSelectorFactory{}
}

// MakeArticleSelector は設定に基づいて適切な ArticleSelector を生成する
func (f *ArticleSelectorFactory) MakeArticleSelector(
	aiConfig *entity.AIConfig,
	promptConfig *entity.PromptConfig,
) (domain.ArticleSelector, error) {
	if aiConfig == nil {
		return nil, fmt.Errorf("ai config is nil")
	}
	if promptConfig == nil {
		return nil, fmt.Errorf("prompt config is nil")
	}

	// Gemini設定がある場合はGemini実装を返す
	if aiConfig.Gemini != nil {
		return newGeminiArticleSelector(aiConfig, promptConfig)
	}

	return nil, fmt.Errorf("no supported AI configuration found")
}
