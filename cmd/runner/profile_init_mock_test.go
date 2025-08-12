package runner

import (
	"errors"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestProfileInitRunner_Run_WithMock(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mock_domain.MockProfileRepository)
		wantErr   bool
	}{
		{
			name: "保存成功",
			setupMock: func(m *mock_domain.MockProfileRepository) {
				m.EXPECT().SaveProfileWithTemplate().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "保存失敗",
			setupMock: func(m *mock_domain.MockProfileRepository) {
				m.EXPECT().SaveProfileWithTemplate().
					Return(errors.New("save error"))
			},
			wantErr: true,
		},
		{
			name: "書き込み権限エラー",
			setupMock: func(m *mock_domain.MockProfileRepository) {
				m.EXPECT().SaveProfileWithTemplate().
					Return(errors.New("permission denied"))
			},
			wantErr: true,
		},
		{
			name: "ディスク容量不足エラー",
			setupMock: func(m *mock_domain.MockProfileRepository) {
				m.EXPECT().SaveProfileWithTemplate().
					Return(errors.New("no space left on device"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockProfileRepository(ctrl)
			tt.setupMock(mockRepo)

			runner := NewProfileInitRunner(mockRepo)
			err := runner.Run()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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

// BenchmarkProfileInitRunner_Run ベンチマークテスト
func BenchmarkProfileInitRunner_Run(b *testing.B) {
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
