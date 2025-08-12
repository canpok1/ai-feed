package runner

import (
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProfileInitRunner_ConcurrentMock(t *testing.T) {
	const goroutines = 5
	results := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockProfileRepository(ctrl)
			mockRepo.EXPECT().SaveProfileWithTemplate().Return(nil)

			runner := NewProfileInitRunner(mockRepo)
			results <- runner.Run()
		}()
	}

	// 全ての並行実行が成功することを確認
	successCount := 0
	for i := 0; i < goroutines; i++ {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	assert.Equal(t, goroutines, successCount, "All concurrent executions should succeed")
}

// BenchmarkProfileInitRunner_ConcurrentRun 並行実行ベンチマークテスト
func BenchmarkProfileInitRunner_ConcurrentRun(b *testing.B) {
	ctrl := gomock.NewController(b)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockProfileRepository(ctrl)
	// 毎回成功するようにセットアップ
	mockRepo.EXPECT().SaveProfileWithTemplate().Return(nil).Times(b.N)

	runner := NewProfileInitRunner(mockRepo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = runner.Run()
	}
}
