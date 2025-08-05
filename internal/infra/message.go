package infra

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// MessageBuilder はテンプレートを使用してメッセージを生成する
type MessageBuilder struct {
	recommendTemplate *template.Template
}

// NewMessageBuilder は新しいMessageBuilderを作成する
func NewMessageBuilder(recommendTemplate string) (*MessageBuilder, error) {
	tmpl, err := template.New("recommend").Parse(recommendTemplate)
	if err != nil {
		return nil, err
	}

	return &MessageBuilder{
		recommendTemplate: tmpl,
	}, nil
}

// templateData はテンプレートで使用するデータ構造
type templateData struct {
	Article      entity.Article
	Comment      *string
	FixedMessage string
}

// BuildRecommendMessage はentity.Recommendとfixed messageを元にメッセージを生成する
func (b *MessageBuilder) BuildRecommendMessage(r *entity.Recommend, fixedMessage string) (string, error) {
	if r == nil {
		return "", fmt.Errorf("recommend cannot be nil")
	}

	data := templateData{
		Article:      r.Article,
		Comment:      r.Comment,
		FixedMessage: fixedMessage,
	}

	var buf bytes.Buffer
	err := b.recommendTemplate.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
