package runner

import (
	"bytes"
	"errors"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/mock_domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewConfigInitRunner(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Given: モックとstderrのセットアップ
	configRepo := mock_domain.NewMockConfigInitRepository(ctrl)
	stderr := &bytes.Buffer{}

	// When: Runnerを作成
	runner := NewConfigInitRunner(configRepo, stderr)

	// Then: Runnerが正常に作成されることを確認
	assert.NotNil(t, runner)
}

func TestConfigInitRunner_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(ctrl *gomock.Controller) *mock_domain.MockConfigInitRepository
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "正常系: 設定ファイルが正常に作成される",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockConfigInitRepository {
				mockRepo := mock_domain.NewMockConfigInitRepository(ctrl)
				mockRepo.EXPECT().SaveWithTemplate().Return(nil)
				return mockRepo
			},
			wantErr: false,
		},
		{
			name: "異常系: 既存ファイルが存在する場合エラー",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockConfigInitRepository {
				mockRepo := mock_domain.NewMockConfigInitRepository(ctrl)
				mockRepo.EXPECT().SaveWithTemplate().Return(errors.New("config file already exists: ./config.yml"))
				return mockRepo
			},
			wantErr:    true,
			wantErrMsg: "failed to create config file",
		},
		{
			name: "異常系: ファイル作成時にエラーが発生",
			setupMock: func(ctrl *gomock.Controller) *mock_domain.MockConfigInitRepository {
				mockRepo := mock_domain.NewMockConfigInitRepository(ctrl)
				mockRepo.EXPECT().SaveWithTemplate().Return(errors.New("permission denied"))
				return mockRepo
			},
			wantErr:    true,
			wantErrMsg: "failed to create config file",
		},
	}

	for _, tt := range tests {
		tt := tt // ループ変数のキャプチャ
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Given: モックとランナーのセットアップ
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := tt.setupMock(ctrl)
			stderr := &bytes.Buffer{}
			runner := NewConfigInitRunner(mockRepo, stderr)
			require.NotNil(t, runner, "runner should not be nil")

			// When: Runメソッドを実行
			err := runner.Run()

			// Then: 期待される結果を検証
			if tt.wantErr {
				assert.ErrorContains(t, err, tt.wantErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// 進行状況メッセージが出力されることを確認
			assert.Contains(t, stderr.String(), "設定テンプレートを生成しています...")
		})
	}
}
