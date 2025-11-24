package comment

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type CommentGeneratorFactory struct{}

func NewCommentGeneratorFactory() domain.CommentGeneratorFactory {
	return &CommentGeneratorFactory{}
}

func (f *CommentGeneratorFactory) MakeCommentGenerator(model *entity.AIConfig, prompt *entity.PromptConfig) (domain.CommentGenerator, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	if prompt == nil {
		return nil, fmt.Errorf("prompt is nil")
	}

	// Gemini設定のチェック
	if model.Gemini == nil {
		return nil, fmt.Errorf("gemini config is nil")
	}
	if model.Gemini.Type == "" {
		return nil, fmt.Errorf("gemini model type is empty")
	}

	// すべてのGeminiモデルをサポート
	// モデルの使用可否判定はGeminiライブラリに任せる
	return newGeminiCommentGenerator(model, prompt, prompt.SystemPrompt)
}
