package message

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
	// 別名記法を既存記法に変換
	converter := entity.NewSlackTemplateAliasConverter()
	convertedTemplate, err := converter.Convert(recommendTemplate)
	if err != nil {
		// 別名変換エラーの場合は、詳細なエラーメッセージを返す
		if aliasErr, ok := err.(*entity.TemplateAliasError); ok {
			return nil, fmt.Errorf("テンプレートエラー: %s", aliasErr.Message)
		}
		return nil, err
	}

	tmpl, err := template.New("recommend").Parse(convertedTemplate)
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
