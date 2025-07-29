package infra

import (
	"fmt"
	"net/http"
	"net/url"

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
func NewMisskeyViewer(instanceURL, accessToken string) (domain.Viewer, error) {
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

// ViewArticles はMisskeyViewerでは未実装です。
func (v *MisskeyViewer) ViewArticles(articles []entity.Article) error {
	return nil
}

// ViewRecommend はMisskeyにノートを投稿します。
func (v *MisskeyViewer) ViewRecommend(recommend *entity.Recommend) error {
	if recommend == nil {
		return fmt.Errorf("recommend is nil")
	}

	text := fmt.Sprintf("Title: %s\nLink: %s", recommend.Article.Title, recommend.Article.Link)
	if recommend.Comment != nil {
		text = fmt.Sprintf("%s\nComment: %s", text, *recommend.Comment)
	}

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
