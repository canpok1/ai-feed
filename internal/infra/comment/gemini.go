package comment

import (
	"context"
	"fmt"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"google.golang.org/genai"
)

type geminiCommentGenerator struct {
	model        *entity.AIConfig
	prompt       *entity.PromptConfig
	systemPrompt string
	client       *genai.Client
}

func newGeminiCommentGenerator(model *entity.AIConfig, prompt *entity.PromptConfig, systemPrompt string) (domain.CommentGenerator, error) {
	// クライアントの初期化はここで行い、構造体に保持する
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  model.Gemini.APIKey,
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
	prompt, err := g.prompt.BuildCommentPrompt(article)
	if err != nil {
		return "", fmt.Errorf("プロンプト生成エラー: %w", err)
	}

	contents := genai.Text(prompt)
	config := genai.GenerateContentConfig{}
	if g.systemPrompt != "" {
		config.SystemInstruction = genai.NewContentFromText(g.systemPrompt, "")
	}

	resp, err := g.client.Models.GenerateContent(ctx, g.model.Gemini.Type, contents, &config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return resp.Text(), nil
}
