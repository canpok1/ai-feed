package domain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/canpok1/ai-feed/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProfileRepository はProfileRepositoryのモック実装
type MockProfileRepository struct {
	mock.Mock
}

func (m *MockProfileRepository) LoadProfile() (*entity.Profile, error) {
	args := m.Called()
	result := args.Get(0)
	if result == nil {
		return nil, args.Error(1)
	}
	return result.(*entity.Profile), args.Error(1)
}

// MockProfileValidator はProfileValidatorのモック実装
type MockProfileValidator struct {
	mock.Mock
}

func (m *MockProfileValidator) Validate(profile *entity.Profile) *ValidationResult {
	args := m.Called(profile)
	return args.Get(0).(*ValidationResult)
}

// TestProfileServiceImpl_ValidateProfile はValidateProfileメソッドをテストする
func TestProfileServiceImpl_ValidateProfile(t *testing.T) {
	tests := []struct {
		name               string
		inputPath          string
		setupFunc          func() (string, func()) // セットアップ関数（パス, クリーンアップ関数）
		mockProfile        *entity.Profile
		mockLoadError      error
		mockValidation     *ValidationResult
		expectedResult     *ValidationResult
		expectedError      string
		expectFileAccess   bool
	}{
		{
			name:      "正常なプロファイルファイルの読み込み",
			inputPath: "valid_profile.yml",
			setupFunc: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "profile_test_")
				require.NoError(t, err)
				profilePath := filepath.Join(tempDir, "valid_profile.yml")
				err = ioutil.WriteFile(profilePath, []byte("test: data"), 0644)
				require.NoError(t, err)
				return profilePath, func() { os.RemoveAll(tempDir) }
			},
			mockProfile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: "test-key"},
				},
			},
			mockValidation: &ValidationResult{
				IsValid:  true,
				Errors:   nil,
				Warnings: nil,
			},
			expectedResult: &ValidationResult{
				IsValid:  true,
				Errors:   nil,
				Warnings: nil,
			},
			expectFileAccess: true,
		},
		{
			name:               "存在しないファイル",
			inputPath:          "non_existent.yml",
			setupFunc:          func() (string, func()) { return "non_existent.yml", func() {} },
			expectedError:      "profile file not found",
			expectFileAccess:   false,
		},
		{
			name:      "ファイル読み込みエラー",
			inputPath: "error_profile.yml",
			setupFunc: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "profile_test_")
				require.NoError(t, err)
				profilePath := filepath.Join(tempDir, "error_profile.yml")
				err = ioutil.WriteFile(profilePath, []byte("test: data"), 0644)
				require.NoError(t, err)
				return profilePath, func() { os.RemoveAll(tempDir) }
			},
			mockLoadError:    fmt.Errorf("failed to parse YAML"),
			expectedError:    "failed to load profile: failed to parse YAML",
			expectFileAccess: true,
		},
		{
			name:      "バリデーションでエラーがある場合",
			inputPath: "invalid_profile.yml",
			setupFunc: func() (string, func()) {
				tempDir, err := ioutil.TempDir("", "profile_test_")
				require.NoError(t, err)
				profilePath := filepath.Join(tempDir, "invalid_profile.yml")
				err = ioutil.WriteFile(profilePath, []byte("test: data"), 0644)
				require.NoError(t, err)
				return profilePath, func() { os.RemoveAll(tempDir) }
			},
			mockProfile: &entity.Profile{
				AI: &entity.AIConfig{
					Gemini: &entity.GeminiConfig{APIKey: ""},
				},
			},
			mockValidation: &ValidationResult{
				IsValid: false,
				Errors:  []string{"Gemini API key is not configured"},
				Warnings: nil,
			},
			expectedResult: &ValidationResult{
				IsValid: false,
				Errors:  []string{"Gemini API key is not configured"},
				Warnings: nil,
			},
			expectFileAccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ
			actualPath, cleanup := tt.setupFunc()
			defer cleanup()

			// モックの準備
			mockValidator := &MockProfileValidator{}
			mockRepoFactory := func(path string) ProfileRepository {
				mockRepo := &MockProfileRepository{}
				if tt.expectFileAccess {
					mockRepo.On("LoadProfile").Return(tt.mockProfile, tt.mockLoadError)
				}
				return mockRepo
			}

			if tt.mockValidation != nil {
				mockValidator.On("Validate", tt.mockProfile).Return(tt.mockValidation)
			}

			// ProfileService作成
			service := NewProfileService(mockValidator, mockRepoFactory)

			// テスト実行
			result, err := service.ValidateProfile(actualPath)

			// アサーション
			if tt.expectedError != "" {
				assert.Error(t, err, "Should return error")
				assert.Contains(t, err.Error(), tt.expectedError, "Error message should contain expected text")
				assert.Nil(t, result, "Result should be nil on error")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.Equal(t, tt.expectedResult, result, "Result should match expected")
			}

			// モックの検証
			mockValidator.AssertExpectations(t)
		})
	}
}

// TestProfileServiceImpl_ResolvePath はResolvePathメソッドをテストする
func TestProfileServiceImpl_ResolvePath(t *testing.T) {
	// 現在のディレクトリとホームディレクトリを取得
	currentDir, err := os.Getwd()
	require.NoError(t, err)

	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	// テスト用環境変数設定
	originalTestVar := os.Getenv("TEST_VAR")
	os.Setenv("TEST_VAR", "test_value")
	defer func() {
		if originalTestVar == "" {
			os.Unsetenv("TEST_VAR")
		} else {
			os.Setenv("TEST_VAR", originalTestVar)
		}
	}()

	tests := []struct {
		name           string
		inputPath      string
		expectedPath   string
		expectedError  string
		setupFunc      func() func() // セットアップ関数（クリーンアップ関数を返す）
	}{
		{
			name:         "絶対パス",
			inputPath:    "/tmp/test.yml",
			expectedPath: "/tmp/test.yml",
		},
		{
			name:         "相対パス",
			inputPath:    "test.yml",
			expectedPath: filepath.Join(currentDir, "test.yml"),
		},
		{
			name:         "現在ディレクトリ指定",
			inputPath:    "./test.yml",
			expectedPath: filepath.Join(currentDir, "test.yml"),
		},
		{
			name:         "ホームディレクトリ展開",
			inputPath:    "~/test.yml",
			expectedPath: filepath.Join(homeDir, "test.yml"),
		},
		{
			name:         "ホームディレクトリ + サブディレクトリ",
			inputPath:    "~/config/test.yml",
			expectedPath: filepath.Join(homeDir, "config", "test.yml"),
		},
		{
			name:         "環境変数展開 - $VAR形式",
			inputPath:    "$TEST_VAR/test.yml",
			expectedPath: filepath.Join(currentDir, "test_value", "test.yml"),
		},
		{
			name:         "環境変数展開 - ${VAR}形式",
			inputPath:    "${TEST_VAR}/test.yml",
			expectedPath: filepath.Join(currentDir, "test_value", "test.yml"),
		},
		{
			name:         "複合パターン - ホーム + 環境変数",
			inputPath:    "~/config/$TEST_VAR/test.yml",
			expectedPath: filepath.Join(homeDir, "config", "test_value", "test.yml"),
		},
		{
			name:         "未定義環境変数",
			inputPath:    "$UNDEFINED_VAR/test.yml",
			expectedPath: "/test.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc()
				defer cleanup()
			}

			// モックサービス作成（ResolvePath用なのでvalidatorとrepoFactoryは最小限）
			service := NewProfileService(&MockProfileValidator{}, func(string) ProfileRepository { return &MockProfileRepository{} })

			// テスト実行
			result, err := service.ResolvePath(tt.inputPath)

			// アサーション
			if tt.expectedError != "" {
				assert.Error(t, err, "Should return error")
				assert.Contains(t, err.Error(), tt.expectedError, "Error message should contain expected text")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.Equal(t, tt.expectedPath, result, "Resolved path should match expected")
			}
		})
	}
}

// TestProfileServiceImpl_ValidateProfile_PathResolution はパス解決と統合したテストを実行する
func TestProfileServiceImpl_ValidateProfile_PathResolution(t *testing.T) {
	// テンポラリディレクトリ作成
	tempDir, err := ioutil.TempDir("", "profile_path_test_")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// テストファイル作成
	profilePath := filepath.Join(tempDir, "test_profile.yml")
	err = ioutil.WriteFile(profilePath, []byte("ai:\n  gemini:\n    api_key: test"), 0644)
	require.NoError(t, err)

	// 環境変数設定
	originalTestDir := os.Getenv("TEST_DIR")
	os.Setenv("TEST_DIR", tempDir)
	defer func() {
		if originalTestDir == "" {
			os.Unsetenv("TEST_DIR")
		} else {
			os.Setenv("TEST_DIR", originalTestDir)
		}
	}()

	tests := []struct {
		name      string
		inputPath string
		setupFunc func() func() // セットアップ関数（クリーンアップ関数を返す）
	}{
		{
			name:      "絶対パス",
			inputPath: profilePath,
		},
		{
			name:      "環境変数を使用したパス",
			inputPath: "$TEST_DIR/test_profile.yml",
		},
		{
			name:      "相対パス",
			inputPath: "./test_profile.yml",
			setupFunc: func() func() {
				originalDir, err := os.Getwd()
				require.NoError(t, err)
				err = os.Chdir(tempDir)
				require.NoError(t, err)
				return func() { os.Chdir(originalDir) }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セットアップ
			var cleanup func()
			if tt.setupFunc != nil {
				cleanup = tt.setupFunc()
				defer cleanup()
			}

			// モック準備
			mockValidator := &MockProfileValidator{}
			mockProfile := &entity.Profile{
				AI: &entity.AIConfig{Gemini: &entity.GeminiConfig{APIKey: "test"}},
			}
			mockResult := &ValidationResult{IsValid: true, Errors: nil, Warnings: nil}

			mockRepoFactory := func(path string) ProfileRepository {
				mockRepo := &MockProfileRepository{}
				mockRepo.On("LoadProfile").Return(mockProfile, nil)
				return mockRepo
			}

			mockValidator.On("Validate", mockProfile).Return(mockResult)

			// サービス作成とテスト実行
			service := NewProfileService(mockValidator, mockRepoFactory)
			result, err := service.ValidateProfile(tt.inputPath)

			// アサーション
			assert.NoError(t, err, "Should not return error")
			assert.Equal(t, mockResult, result, "Result should match expected")

			// モック検証
			mockValidator.AssertExpectations(t)
		})
	}
}

// TestProfileServiceImpl_ErrorHandling はエラーハンドリングの詳細をテストする
func TestProfileServiceImpl_ErrorHandling(t *testing.T) {
	tests := []struct {
		name                string
		pathResolveError    bool
		fileExists          bool
		fileReadable        bool
		repositoryError     error
		expectedErrorPrefix string
	}{
		{
			name:                "ファイルが存在しない",
			fileExists:          false,
			expectedErrorPrefix: "profile file not found",
		},
		{
			name:                "リポジトリエラー",
			fileExists:          true,
			fileReadable:        true,
			repositoryError:     fmt.Errorf("YAML parse error"),
			expectedErrorPrefix: "failed to load profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// テンポラリディレクトリ作成
			tempDir, err := ioutil.TempDir("", "error_test_")
			require.NoError(t, err)
			defer os.RemoveAll(tempDir)

			var testPath string
			if tt.fileExists {
				testPath = filepath.Join(tempDir, "test.yml")
				err = ioutil.WriteFile(testPath, []byte("test: data"), 0644)
				require.NoError(t, err)
			} else {
				testPath = filepath.Join(tempDir, "non_existent.yml")
			}

			// モック準備
			mockValidator := &MockProfileValidator{}
			mockRepoFactory := func(path string) ProfileRepository {
				mockRepo := &MockProfileRepository{}
				if tt.repositoryError != nil {
					mockRepo.On("LoadProfile").Return(nil, tt.repositoryError)
				} else if tt.fileExists {
					// ファイルが存在し読み取り可能な場合はモックを設定
					mockProfile := &entity.Profile{}
					mockRepo.On("LoadProfile").Return(mockProfile, nil)
				}
				return mockRepo
			}

			// サービス作成とテスト実行
			service := NewProfileService(mockValidator, mockRepoFactory)
			result, err := service.ValidateProfile(testPath)

			// アサーション
			assert.Error(t, err, "Should return error")
			if !strings.Contains(err.Error(), tt.expectedErrorPrefix) {
				t.Logf("Expected error prefix: %s", tt.expectedErrorPrefix)
				t.Logf("Actual error: %s", err.Error())
			}
			assert.Contains(t, err.Error(), tt.expectedErrorPrefix, "Error should contain expected prefix")
			assert.Nil(t, result, "Result should be nil on error")
		})
	}
}