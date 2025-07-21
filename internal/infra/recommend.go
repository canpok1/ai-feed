package infra

import (
	"context"
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"google.golang.org/genai"
)

type factoryFunc func(*entity.AIModelConfig, *entity.PromptConfig, string) (domain.CommentGenerator, error)

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

func (f *CommentGeneratorFactory) MakeCommentGenerator(model *entity.AIModelConfig, prompt *entity.PromptConfig, systemPrompt string) (domain.CommentGenerator, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}
	if prompt == nil {
		return nil, fmt.Errorf("prompt is nil")
	}

	if f, ok := f.factoryFuncMap[model.Type]; ok {
		return f(model, prompt, systemPrompt)
	}
	return nil, fmt.Errorf("unsupported model type: %s", model.Type)
}

type geminiCommentGenerator struct {
	model        *entity.AIModelConfig
	prompt       *entity.PromptConfig
	systemPrompt string
	client       *genai.Client
}

func newGeminiCommentGenerator(model *entity.AIModelConfig, prompt *entity.PromptConfig, systemPrompt string) (domain.CommentGenerator, error) {
	// クライアントの初期化はここで行い、構造体に保持する
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  model.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &geminiCommentGenerator{
		model:        model,
		prompt:       prompt,
		client:       client,
		systemPrompt: systemPrompt,
	}, nil
}

func (g *geminiCommentGenerator) Generate(ctx context.Context, article *entity.Article) (string, error) {
	contents := genai.Text(g.prompt.MakeCommentPromptTemplate(article))
	config := genai.GenerateContentConfig{}
	if g.systemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(g.systemPrompt, "")
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model.Type, contents, &config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return resp.Text(), nil
}
