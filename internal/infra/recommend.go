package infra

import (
	"context"
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"google.golang.org/genai"
)

type factoryFunc func(*entity.AIModelConfig, *entity.PromptConfig) domain.CommentGenerator

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

func (f *CommentGeneratorFactory) MakeCommentGenerator(model *entity.AIModelConfig, prompt *entity.PromptConfig) (domain.CommentGenerator, error) {
	if f, ok := f.factoryFuncMap[model.Type]; ok {
		return f(model, prompt), nil
	}
	return nil, fmt.Errorf("unsupported model type: %s", model.Type)
}

type geminiCommentGenerator struct {
	model  *entity.AIModelConfig
	prompt *entity.PromptConfig
}

func newGeminiCommentGenerator(model *entity.AIModelConfig, prompt *entity.PromptConfig) domain.CommentGenerator {
	return &geminiCommentGenerator{
		model:  model,
		prompt: prompt,
	}
}

func (g *geminiCommentGenerator) Generate(ctx context.Context, article *entity.Article) (string, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  g.model.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create gemini client: %w", err)
	}

	prompt := genai.Text(g.prompt.MakeCommentPromptTemplate(article))
	resp, err := client.Models.GenerateContent(ctx, g.model.Type, prompt, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return resp.Text(), nil
}
