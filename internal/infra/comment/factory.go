package comment

import (
	"fmt"
	"strings"

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

	// Gemini設定のバリデーション
	if model.Gemini == nil {
		return nil, fmt.Errorf("gemini config is nil")
	}
	if result := model.Gemini.Validate(); !result.IsValid {
		return nil, fmt.Errorf("invalid gemini config: %s", strings.Join(result.Errors, "; "))
	}

	// すべてのGeminiモデルをサポート
	// モデルの使用可否判定はGeminiライブラリに任せる
	return newGeminiCommentGenerator(model, prompt, prompt.SystemPrompt)
}
