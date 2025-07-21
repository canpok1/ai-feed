package infra

import (
	"context"
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
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

func (g *geminiCommentGenerator) Generate(article *entity.Article) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(g.model.APIKey))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(g.model.Type)

	prompt := genai.Text(g.prompt.MakeCommentPromptTemplate(article))
	resp, err := model.GenerateContent(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	comment, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("generated content is not text")
	}

	return string(comment), nil
}
