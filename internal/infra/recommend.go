package infra

import (
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

type factoryFunc func(entity.AIModelConfig, entity.PromptConfig) domain.CommentGenerator

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

func (f *CommentGeneratorFactory) MakeCommentGenerator(model entity.AIModelConfig, prompt entity.PromptConfig) (domain.CommentGenerator, error) {
	if f, ok := f.factoryFuncMap[model.Type]; ok {
		return f(model, prompt), nil
	}
	return nil, fmt.Errorf("unsupported model type: %s", model.Type)
}

type geminiCommentGenerator struct {
	model  entity.AIModelConfig
	prompt entity.PromptConfig
}

func newGeminiCommentGenerator(model entity.AIModelConfig, prompt entity.PromptConfig) domain.CommentGenerator {
	return &geminiCommentGenerator{
		model:  model,
		prompt: prompt,
	}
}

func (g *geminiCommentGenerator) Generate(article entity.Article) (string, error) {
	// TODO 実装
	return "生成したコメント", nil
}
