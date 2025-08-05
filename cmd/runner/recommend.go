package runner

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/infra"
)

// ErrNoArticlesFound は記事が見つからなかった場合のsentinel error
var ErrNoArticlesFound = errors.New("no articles found in the feed")

// RecommendParams はrecommendコマンドの実行パラメータを表す構造体
type RecommendParams struct {
	URLs []string
}

// RecommendRunner はrecommendコマンドのビジネスロジックを実行する構造体
type RecommendRunner struct {
	fetcher     *domain.Fetcher
	recommender domain.Recommender
	viewers     []domain.Viewer
}

// NewRecommendRunner はRecommendRunnerの新しいインスタンスを作成する
func NewRecommendRunner(fetchClient domain.FetchClient, recommender domain.Recommender, stdout io.Writer, stderr io.Writer, outputConfig *infra.OutputConfig, promptConfig *infra.PromptConfig) (*RecommendRunner, error) {
	fetcher := domain.NewFetcher(
		fetchClient,
		func(url string, err error) error {
			fmt.Fprintf(stderr, "Error fetching feed from %s: %v\n", url, err)
			return err
		},
	)
	viewer, err := infra.NewStdViewer(stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create viewer: %w", err)
	}
	viewers := []domain.Viewer{viewer}

	if outputConfig != nil {
		if outputConfig.SlackAPI != nil {
			slackViewer := infra.NewSlackViewer(outputConfig.SlackAPI.ToEntity())
			viewers = append(viewers, slackViewer)
		}
		if outputConfig.Misskey != nil {
			misskeyViewer, err := infra.NewMisskeyViewer(outputConfig.Misskey.APIURL, outputConfig.Misskey.APIToken)
			if err != nil {
				return nil, fmt.Errorf("failed to create Misskey viewer: %w", err)
			}
			viewers = append(viewers, misskeyViewer)
		}
	}

	return &RecommendRunner{
		fetcher:     fetcher,
		recommender: recommender,
		viewers:     viewers,
	}, nil
}

// Run はrecommendコマンドのビジネスロジックを実行する
func (r *RecommendRunner) Run(ctx context.Context, params *RecommendParams, profile infra.Profile) error {
	allArticles, err := r.fetcher.Fetch(params.URLs, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch articles: %w", err)
	}

	if len(allArticles) == 0 {
		return ErrNoArticlesFound
	}

	if profile.AI == nil || profile.Prompt == nil {
		return fmt.Errorf("AI model or prompt is not configured")
	}

	aiConfigEntity := profile.AI.ToEntity()
	promptConfigEntity := profile.Prompt.ToEntity()

	recommend, err := r.recommender.Recommend(
		ctx,
		aiConfigEntity,
		promptConfigEntity,
		allArticles)
	if err != nil {
		return fmt.Errorf("failed to recommend article: %w", err)
	}

	var errs []error
	for _, viewer := range r.viewers {
		if viewErr := viewer.ViewRecommend(recommend, profile.Prompt.FixedMessage); viewErr != nil {
			errs = append(errs, fmt.Errorf("failed to view recommend: %w", viewErr))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to view all recommends: %v", errs)
	}

	return nil
}
