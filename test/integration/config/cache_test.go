//go:build integration

// Package config はキャッシュ設定の統合テストを提供する
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canpok1/ai-feed/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheConfig_EnabledDefault はcache.enabledのデフォルト値がfalseであることを検証する
// enabled が省略された場合、デフォルトで無効（false）になること
func TestCacheConfig_EnabledDefault(t *testing.T) {
	// Enabledが省略されたCacheConfig
	cacheConfig := &infra.CacheConfig{
		Enabled:       nil, // 省略
		FilePath:      "/tmp/test-cache.jsonl",
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// デフォルト値がfalseであることを確認
	assert.False(t, entityCache.Enabled,
		"enabledが省略された場合、デフォルト値はfalseになるはずです")
}

// TestCacheConfig_FilePathDefault はcache.file_pathのデフォルト値を検証する
// file_path が省略された場合、~/.ai-feed/recommend_history.jsonl に設定されること
func TestCacheConfig_FilePathDefault(t *testing.T) {
	// FilePathが省略されたCacheConfig
	enabled := true
	cacheConfig := &infra.CacheConfig{
		Enabled:       &enabled,
		FilePath:      "", // 省略
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "ホームディレクトリの取得に成功するはずです")

	// デフォルトのパスが正しく設定されることを確認
	expectedPath := filepath.Join(homeDir, ".ai-feed", "recommend_history.jsonl")
	assert.Equal(t, expectedPath, entityCache.FilePath,
		"file_pathが省略された場合、デフォルト値は ~/.ai-feed/recommend_history.jsonl（展開後）になるはずです")
}

// TestCacheConfig_MaxEntriesDefault はcache.max_entriesのデフォルト値が1000であることを検証する
// max_entries が省略または0以下の場合、デフォルトで1000になること
func TestCacheConfig_MaxEntriesDefault(t *testing.T) {
	tests := []struct {
		name       string
		maxEntries int
		want       int
	}{
		{
			name:       "max_entries=0の場合、デフォルト値1000が適用される",
			maxEntries: 0,
			want:       1000,
		},
		{
			name:       "max_entries=-1の場合、デフォルト値1000が適用される",
			maxEntries: -1,
			want:       1000,
		},
		{
			name:       "max_entries=500の場合、指定値500が使用される",
			maxEntries: 500,
			want:       500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := true
			cacheConfig := &infra.CacheConfig{
				Enabled:       &enabled,
				FilePath:      "/tmp/test-cache.jsonl",
				MaxEntries:    tt.maxEntries,
				RetentionDays: 7,
			}

			// CacheConfig から entity.CacheConfig に変換
			entityCache, err := cacheConfig.ToEntity()
			require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

			assert.Equal(t, tt.want, entityCache.MaxEntries)
		})
	}
}

// TestCacheConfig_RetentionDaysDefault はcache.retention_daysのデフォルト値が30であることを検証する
// retention_days が省略または0以下の場合、デフォルトで30になること
func TestCacheConfig_RetentionDaysDefault(t *testing.T) {
	tests := []struct {
		name          string
		retentionDays int
		want          int
	}{
		{
			name:          "retention_days=0の場合、デフォルト値30が適用される",
			retentionDays: 0,
			want:          30,
		},
		{
			name:          "retention_days=-1の場合、デフォルト値30が適用される",
			retentionDays: -1,
			want:          30,
		},
		{
			name:          "retention_days=14の場合、指定値14が使用される",
			retentionDays: 14,
			want:          14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enabled := true
			cacheConfig := &infra.CacheConfig{
				Enabled:       &enabled,
				FilePath:      "/tmp/test-cache.jsonl",
				MaxEntries:    100,
				RetentionDays: tt.retentionDays,
			}

			// CacheConfig から entity.CacheConfig に変換
			entityCache, err := cacheConfig.ToEntity()
			require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

			assert.Equal(t, tt.want, entityCache.RetentionDays)
		})
	}
}

// TestCacheConfig_TildeExpansion はチルダ記号がホームディレクトリに展開されることを検証する
// file_path にチルダ（~）が含まれる場合、実際のホームディレクトリに展開されること
func TestCacheConfig_TildeExpansion(t *testing.T) {
	enabled := true
	cacheConfig := &infra.CacheConfig{
		Enabled:       &enabled,
		FilePath:      "~/custom/cache.jsonl", // チルダを含むパス
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "ホームディレクトリの取得に成功するはずです")

	// チルダがホームディレクトリに展開されることを確認
	expectedPath := filepath.Join(homeDir, "custom", "cache.jsonl")
	assert.Equal(t, expectedPath, entityCache.FilePath,
		"チルダは実際のホームディレクトリに展開されるはずです")

	// 結果のパスがチルダを含まないことを確認
	assert.NotContains(t, entityCache.FilePath, "~",
		"展開後のパスにはチルダが含まれないはずです")

	// 結果のパスが絶対パスであることを確認
	assert.True(t, filepath.IsAbs(entityCache.FilePath),
		"展開後のパスは絶対パスであるはずです")
}

// TestCacheConfig_RelativePathConversion は相対パスが絶対パスに変換されることを検証する
// file_path に相対パスが指定された場合、絶対パスに変換されること
func TestCacheConfig_RelativePathConversion(t *testing.T) {
	enabled := true
	cacheConfig := &infra.CacheConfig{
		Enabled:       &enabled,
		FilePath:      "relative/path/cache.jsonl", // 相対パス
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 結果のパスが絶対パスであることを確認
	assert.True(t, filepath.IsAbs(entityCache.FilePath),
		"相対パスは絶対パスに変換されるはずです")

	// パスの末尾が元の相対パスを含むことを確認
	assert.Contains(t, entityCache.FilePath, "relative/path/cache.jsonl",
		"変換後のパスには元のパス構造が含まれるはずです")
}

// TestCacheConfig_AbsolutePathPreserved は絶対パスがそのまま保持されることを検証する
// file_path に絶対パスが指定された場合、そのまま使用されること
func TestCacheConfig_AbsolutePathPreserved(t *testing.T) {
	enabled := true
	absolutePath := "/absolute/path/to/cache.jsonl"
	cacheConfig := &infra.CacheConfig{
		Enabled:       &enabled,
		FilePath:      absolutePath,
		MaxEntries:    100,
		RetentionDays: 7,
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 絶対パスがそのまま保持されることを確認
	assert.Equal(t, absolutePath, entityCache.FilePath,
		"絶対パスはそのまま保持されるはずです")
}

// TestCacheConfig_ExplicitOverridesDefaults は明示的な設定でデフォルト値が上書きされることを検証する
// すべてのフィールドを明示的に設定した場合、デフォルト値ではなく指定値が使用されること
func TestCacheConfig_ExplicitOverridesDefaults(t *testing.T) {
	enabled := true
	cacheConfig := &infra.CacheConfig{
		Enabled:       &enabled,
		FilePath:      "/custom/path/cache.jsonl",
		MaxEntries:    500, // デフォルト1000ではない
		RetentionDays: 14,  // デフォルト30ではない
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// 明示的に設定した値が使用されることを確認
	assert.True(t, entityCache.Enabled,
		"明示的にtrueを設定した場合、enabledはtrueになるはずです")
	assert.Equal(t, "/custom/path/cache.jsonl", entityCache.FilePath,
		"明示的に設定したfile_pathが使用されるはずです")
	assert.Equal(t, 500, entityCache.MaxEntries,
		"明示的に設定したmax_entriesが使用されるはずです")
	assert.Equal(t, 14, entityCache.RetentionDays,
		"明示的に設定したretention_daysが使用されるはずです")
}

// TestCacheConfig_NilConfig はCacheConfigがnilの場合の動作を検証する
// CacheConfig がnilの場合、ToEntity()はnilを返すこと
func TestCacheConfig_NilConfig(t *testing.T) {
	var cacheConfig *infra.CacheConfig = nil

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "nilのCacheConfigでToEntity()はエラーを返さないはずです")

	// nilが返されることを確認
	assert.Nil(t, entityCache,
		"nilのCacheConfigはnilのentity.CacheConfigに変換されるはずです")
}

// TestCacheConfig_AllDefaults は全てのフィールドがデフォルト値を持つことを検証する
// すべてのフィールドが省略・ゼロ値の場合、適切なデフォルト値が適用されること
func TestCacheConfig_AllDefaults(t *testing.T) {
	// すべてのフィールドがゼロ値のCacheConfig
	cacheConfig := &infra.CacheConfig{
		Enabled:       nil, // デフォルト: false
		FilePath:      "",  // デフォルト: ~/.ai-feed/recommend_history.jsonl
		MaxEntries:    0,   // デフォルト: 1000
		RetentionDays: 0,   // デフォルト: 30
	}

	// CacheConfig から entity.CacheConfig に変換
	entityCache, err := cacheConfig.ToEntity()
	require.NoError(t, err, "ToEntity()はエラーを返さないはずです")

	// ホームディレクトリを取得
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "ホームディレクトリの取得に成功するはずです")

	// すべてのデフォルト値が適用されることを確認
	assert.False(t, entityCache.Enabled,
		"enabledのデフォルト値はfalseです")
	expectedPath := filepath.Join(homeDir, ".ai-feed", "recommend_history.jsonl")
	assert.Equal(t, expectedPath, entityCache.FilePath,
		"file_pathのデフォルト値は ~/.ai-feed/recommend_history.jsonl（展開後）です")
	assert.Equal(t, 1000, entityCache.MaxEntries,
		"max_entriesのデフォルト値は1000です")
	assert.Equal(t, 30, entityCache.RetentionDays,
		"retention_daysのデフォルト値は30です")
}
