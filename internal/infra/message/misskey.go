package message

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"text/template"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/yitsushi/go-misskey"
	"github.com/yitsushi/go-misskey/models"
	"github.com/yitsushi/go-misskey/services/notes"
)

// MisskeyTemplateData はMisskeyメッセージテンプレートで使用するデータ
type MisskeyTemplateData struct {
	Article      *entity.Article
	Comment      *string
	FixedMessage string
}

// DefaultMisskeyMessageTemplate はデフォルトのMisskeyメッセージテンプレート
const DefaultMisskeyMessageTemplate = `{{if .Comment}}{{.Comment}}
{{end}}{{.Article.Title}}
{{.Article.Link}}{{if .FixedMessage}}
{{.FixedMessage}}{{end}}`

// MisskeySender はMisskey APIと通信するためのクライアントです。

type MisskeySender struct {
	client *misskey.Client
	tmpl   *template.Template
}

// NewMisskeySender は新しいMisskeySenderのインスタンスを作成します。
func NewMisskeySender(instanceURL, accessToken string, messageTemplate *string) (domain.MessageSender, error) {
	parsedURL, err := url.Parse(instanceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse instance URL: %w", err)
	}

	client, err := misskey.NewClientWithOptions(
		misskey.WithBaseURL(parsedURL.Scheme, parsedURL.Host, parsedURL.Path),
		misskey.WithAPIToken(accessToken),
		misskey.WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Misskey client: %w", err)
	}

	// テンプレートの設定
	var templateStr string
	if messageTemplate != nil && strings.TrimSpace(*messageTemplate) != "" {
		templateStr = *messageTemplate
	} else {
		templateStr = DefaultMisskeyMessageTemplate
	}

	// 設定読み込み時にテンプレートは検証済みのため、template.Mustが安全に使用できる
	tmpl := template.Must(template.New("misskey_message").Parse(templateStr))

	return &MisskeySender{
		client: client,
		tmpl:   tmpl,
	}, nil
}

// SendRecommend はMisskeyにノートを投稿します。
func (v *MisskeySender) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	if recommend == nil {
		return fmt.Errorf("recommend is nil")
	}

	// テンプレートデータを作成
	templateData := &MisskeyTemplateData{
		Article:      &recommend.Article,
		Comment:      recommend.Comment,
		FixedMessage: fixedMessage,
	}

	// パース済みテンプレートを直接実行
	var buf bytes.Buffer
	if err := v.tmpl.Execute(&buf, templateData); err != nil {
		return err
	}

	text := buf.String()

	params := notes.CreateRequest{
		Text:       &text,
		Visibility: models.VisibilityPublic,
	}

	_, err := v.client.Notes().Create(params)
	if err != nil {
		return fmt.Errorf("failed to create note: %w", err)
	}

	return nil
}
