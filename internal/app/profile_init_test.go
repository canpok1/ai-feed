package app

import (
	"bytes"
	"errors"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewProfileInitRunner(t *testing.T) {
	t.Parallel()

	t.Run("正常系: Runnerが正常に作成される", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := mock_domain.NewMockProfileTemplateRepository(ctrl)
		stderr := &bytes.Buffer{}

		runner, err := NewProfileInitRunner(mockRepo, stderr)

		assert.NoError(t, err)
		assert.NotNil(t, runner)
	})

	t.Run("異常系: templateRepoがnilの場合はエラー", func(t *testing.T) {
		t.Parallel()
		stderr := &bytes.Buffer{}

		runner, err := NewProfileInitRunner(nil, stderr)

		assert.Error(t, err)
		assert.Nil(t, runner)
		assert.Contains(t, err.Error(), "templateRepo cannot be nil")
	})

	t.Run("異常系: stderrがnilの場合はエラー", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		mockRepo := mock_domain.NewMockProfileTemplateRepository(ctrl)

		runner, err := NewProfileInitRunner(mockRepo, nil)

		assert.Error(t, err)
		assert.Nil(t, runner)
		assert.Contains(t, err.Error(), "stderr cannot be nil")
	})
}

func TestProfileInitRunner_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(ctrl *gomock.Controller) *mock_domain.MockProfileTemplateRepository
		wantErr   bool
		errMsg    string
	}{
		{
			name: "正常系: テンプレート保存成功",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockProfileTemplateRepository {
				mockRepo := mock_domain.NewMockProfileTemplateRepository(ctrl)
				mockRepo.EXPECT().SaveProfileTemplate().Return(nil)
				return mockRepo
			},
			wantErr: false,
		},
		{
			name: "異常系: テンプレート保存失敗",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockProfileTemplateRepository {
				mockRepo := mock_domain.NewMockProfileTemplateRepository(ctrl)
				mockRepo.EXPECT().SaveProfileTemplate().Return(errors.New("save failed"))
				return mockRepo
			},
			wantErr: true,
			errMsg:  "failed to create profile file",
		},
		{
			name: "異常系: ファイルが既に存在する",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockProfileTemplateRepository {
				mockRepo := mock_domain.NewMockProfileTemplateRepository(ctrl)
				mockRepo.EXPECT().SaveProfileTemplate().Return(errors.New("profile file already exists"))
				return mockRepo
			},
			wantErr: true,
			errMsg:  "failed to create profile file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			mockRepo := tt.setupMock(ctrl)
			stderr := &bytes.Buffer{}

			runner, runnerErr := NewProfileInitRunner(mockRepo, stderr)
			require.NoError(t, runnerErr)
			err := runner.Run()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// 進行状況メッセージがstderrに出力されていることを確認
			assert.Contains(t, stderr.String(), "設定テンプレートを生成しています...")
		})
	}
}
