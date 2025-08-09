package infra

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/yitsushi/go-misskey"
	"github.com/yitsushi/go-misskey/models"
	"github.com/yitsushi/go-misskey/services/notes"
)

// MisskeyViewer はMisskey APIと通信するためのクライアントです。
type MisskeyViewer struct {
	client *misskey.Client
}

// NewMisskeyViewer は新しいMisskeyViewerのインスタンスを作成します。
func NewMisskeyViewer(instanceURL, accessToken string) (domain.MessageSender, error) {
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

	return &MisskeyViewer{client: client}, nil
}

// SendArticles はMisskeyViewerでは未実装です。
func (v *MisskeyViewer) SendArticles(articles []entity.Article) error {
	return nil
}

// SendRecommend はMisskeyにノートを投稿します。
func (v *MisskeyViewer) SendRecommend(recommend *entity.Recommend, fixedMessage string) error {
	if recommend == nil {
		return fmt.Errorf("recommend is nil")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\nLink: %s", recommend.Article.Title, recommend.Article.Link)
	if recommend.Comment != nil {
		fmt.Fprintf(&b, "\nComment: %s", *recommend.Comment)
	}
	// fixedMessage を追加
	if fixedMessage != "" {
		fmt.Fprintf(&b, "\nFixed Message: %s", fixedMessage)
	}

	text := b.String()

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
