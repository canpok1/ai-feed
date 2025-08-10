package comment

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type factoryFunc func(*entity.AIConfig, *entity.PromptConfig, string) (domain.CommentGenerator, error)

type CommentGeneratorFactory struct {
	factoryFuncMap map[string]factoryFunc
}

func NewCommentGeneratorFactory() domain.CommentGeneratorFactory {
	return &CommentGeneratorFactory{
		factoryFuncMap: map[string]factoryFunc{
			"gemini-2.5-flash": newGeminiCommentGenerator,
			"gemini-2.5-pro":   newGeminiCommentGenerator,
		},
	}
}

func (f *CommentGeneratorFactory) MakeCommentGenerator(model *entity.AIConfig, prompt *entity.PromptConfig) (domain.CommentGenerator, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	if prompt == nil {
		return nil, fmt.Errorf("prompt is nil")
	}

	if f, ok := f.factoryFuncMap[model.Gemini.Type]; ok {
		return f(model, prompt, prompt.SystemPrompt)
	}
	return nil, fmt.Errorf("unsupported model type: %s", model.Gemini.Type)
}
