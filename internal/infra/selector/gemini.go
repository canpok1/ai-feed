package selector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"google.golang.org/genai"
)

// geminiArticleSelector はGemini APIを使用した記事選択の実装
type geminiArticleSelector struct {
	client       *genai.Client
	modelType    string
	systemPrompt string
	prompt       string
}

// newGeminiArticleSelector は新しいgeminiArticleSelectorを作成する
func newGeminiArticleSelector(
	aiConfig *entity.AIConfig,
	promptConfig *entity.PromptConfig,
) (domain.ArticleSelector, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  aiConfig.Gemini.APIKey.Value(),
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &geminiArticleSelector{
		client:       client,
		modelType:    aiConfig.Gemini.Type,
		systemPrompt: promptConfig.SystemPrompt,
		prompt:       promptConfig.SelectorPrompt,
	}, nil
}

func (g *geminiArticleSelector) Select(ctx context.Context, articles []entity.Article) (*entity.Article, error) {
	if len(articles) == 0 {
		return nil, fmt.Errorf("no articles provided")
	}

	// プロンプト生成
	prompt := g.buildSelectionPrompt(articles)

	// Gemini APIに送信（構造化出力）
	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(g.systemPrompt, ""),
		ResponseMIMEType:  "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"selected_index": {
					Type:        genai.TypeInteger,
					Description: "選択した記事のインデックス（0始まり）",
				},
			},
			Required: []string{"selected_index"},
		},
	}
	resp, err := g.client.Models.GenerateContent(ctx, g.modelType, genai.Text(prompt), config)

	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// レスポンスパース
	var result struct {
		SelectedIndex int `json:"selected_index"`
	}
	if err := json.Unmarshal([]byte(resp.Text()), &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// バリデーション
	if result.SelectedIndex < 0 || result.SelectedIndex >= len(articles) {
		return nil, fmt.Errorf("invalid index: %d (total articles: %d)", result.SelectedIndex, len(articles))
	}

	return &articles[result.SelectedIndex], nil
}

// buildSelectionPrompt は記事選択用のプロンプトを生成する
func (g *geminiArticleSelector) buildSelectionPrompt(articles []entity.Article) string {
	var sb strings.Builder

	// プロンプトが設定されていればそれを使用
	if g.prompt != "" {
		sb.WriteString(g.prompt)
		sb.WriteString("\n\n")
	}

	// 記事リストを追加
	for i, article := range articles {
		sb.WriteString(fmt.Sprintf("[%d] タイトル: %s\n", i, article.Title))
		sb.WriteString(fmt.Sprintf("URL: %s\n", article.Link))
		sb.WriteString(fmt.Sprintf("内容: %s\n\n", article.Content))
	}

	return sb.String()
}
