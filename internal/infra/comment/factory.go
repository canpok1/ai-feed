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

	// すべてのGeminiモデルをサポート
	// モデルの使用可否判定はGeminiライブラリに任せる
	return newGeminiCommentGenerator(model, prompt, prompt.SystemPrompt)
}
