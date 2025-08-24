package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

func TestNewFileRecommendCache(t *testing.T) {
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      "/tmp/test_cache.jsonl",
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache := NewFileRecommendCache(config)

	if cache.filePath != config.FilePath {
		t.Errorf("Expected filePath %s, got %s", config.FilePath, cache.filePath)
	}
	if cache.lockPath != config.FilePath+".lock" {
		t.Errorf("Expected lockPath %s.lock, got %s", config.FilePath, cache.lockPath)
	}
	if cache.config != config {
		t.Error("Config not set correctly")
	}
	if cache.urlSet == nil {
		t.Error("urlSet not initialized")
	}
	if cache.entries == nil {
		t.Error("entries not initialized")
	}
}

func TestFileRecommendCache_Initialize(t *testing.T) {
	t.Run("初期化成功", func(t *testing.T) {
		tmpDir := t.TempDir()
		config := &entity.CacheConfig{
			Enabled:       true,
			FilePath:      filepath.Join(tmpDir, "cache.jsonl"),
			MaxEntries:    100,
			RetentionDays: 7,
		}

		cache := NewFileRecommendCache(config)
		err := cache.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		// クリーンアップ
		cache.Close()
	})

	t.Run("無効化されたキャッシュ", func(t *testing.T) {
		config := &entity.CacheConfig{
			Enabled:       false,
			FilePath:      "/tmp/disabled_cache.jsonl",
			MaxEntries:    100,
			RetentionDays: 7,
		}

		cache := NewFileRecommendCache(config)
		err := cache.Initialize()
		if err != nil {
			t.Fatalf("Initialize should succeed for disabled cache: %v", err)
		}
	})

	t.Run("既存ファイルの読み込み", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheFile := filepath.Join(tmpDir, "existing_cache.jsonl")
		config := &entity.CacheConfig{
			Enabled:       true,
			FilePath:      cacheFile,
			MaxEntries:    100,
			RetentionDays: 7,
		}

		// 既存のキャッシュファイルを作成
		entries := []domain.RecommendEntry{
			{URL: "https://example.com/1", Title: "Test 1", PostedAt: time.Now().AddDate(0, 0, -1)},
			{URL: "https://example.com/2", Title: "Test 2", PostedAt: time.Now().AddDate(0, 0, -2)},
		}

		file, err := os.Create(cacheFile)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		for _, entry := range entries {
			data, _ := json.Marshal(entry)
			file.WriteString(string(data) + "\n")
		}
		file.Close()

		cache := NewFileRecommendCache(config)
		err = cache.Initialize()
		if err != nil {
			t.Fatalf("Initialize failed: %v", err)
		}

		if len(cache.entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(cache.entries))
		}
		if len(cache.urlSet) != 2 {
			t.Errorf("Expected 2 URLs in urlSet, got %d", len(cache.urlSet))
		}

		cache.Close()
	})
}

func TestFileRecommendCache_IsCached(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "cache.jsonl"),
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()
	defer cache.Close()

	t.Run("キャッシュされていないURL", func(t *testing.T) {
		if cache.IsCached("https://example.com/not-cached") {
			t.Error("URL should not be cached")
		}
	})

	t.Run("キャッシュされたURL", func(t *testing.T) {
		cache.AddEntry("https://example.com/cached", "Test Article")
		if !cache.IsCached("https://example.com/cached") {
			t.Error("URL should be cached")
		}
	})

	t.Run("URL正規化", func(t *testing.T) {
		cache.AddEntry("https://example.com/test/", "Test Article")
		if !cache.IsCached("https://example.com/test") {
			t.Error("URL normalization should work (trailing slash)")
		}
	})

	t.Run("無効化されたキャッシュ", func(t *testing.T) {
		disabledConfig := &entity.CacheConfig{
			Enabled:       false,
			FilePath:      filepath.Join(tmpDir, "disabled.jsonl"),
			MaxEntries:    100,
			RetentionDays: 7,
		}
		disabledCache := NewFileRecommendCache(disabledConfig)
		disabledCache.Initialize()

		if disabledCache.IsCached("https://example.com/any") {
			t.Error("Disabled cache should always return false")
		}
	})
}

func TestFileRecommendCache_AddEntry(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "cache.jsonl"),
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()
	defer cache.Close()

	t.Run("新しいエントリの追加", func(t *testing.T) {
		err := cache.AddEntry("https://example.com/new", "New Article")
		if err != nil {
			t.Fatalf("AddEntry failed: %v", err)
		}

		if !cache.IsCached("https://example.com/new") {
			t.Error("Added entry should be cached")
		}
		if len(cache.entries) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(cache.entries))
		}
	})

	t.Run("重複エントリの追加", func(t *testing.T) {
		initialCount := len(cache.entries)
		err := cache.AddEntry("https://example.com/new", "Same Article")
		if err != nil {
			t.Fatalf("AddEntry failed: %v", err)
		}

		if len(cache.entries) != initialCount {
			t.Error("Duplicate entry should not be added")
		}
	})

	t.Run("無効化されたキャッシュ", func(t *testing.T) {
		disabledConfig := &entity.CacheConfig{
			Enabled:       false,
			FilePath:      filepath.Join(tmpDir, "disabled.jsonl"),
			MaxEntries:    100,
			RetentionDays: 7,
		}
		disabledCache := NewFileRecommendCache(disabledConfig)
		disabledCache.Initialize()

		err := disabledCache.AddEntry("https://example.com/disabled", "Test")
		if err != nil {
			t.Fatalf("AddEntry should not fail for disabled cache: %v", err)
		}
	})
}

func TestFileRecommendCache_Close(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "cache.jsonl"),
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()

	// ロックファイルが存在することを確認
	if _, err := os.Stat(cache.lockPath); os.IsNotExist(err) {
		t.Error("Lock file should exist after Initialize")
	}

	err := cache.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// ロックファイルが削除されていることを確認
	if _, err := os.Stat(cache.lockPath); !os.IsNotExist(err) {
		t.Error("Lock file should be removed after Close")
	}
}

func TestFileRecommendCache_saveToFile(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "save_test.jsonl"),
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()
	defer cache.Close()

	// テストデータを追加
	testEntries := []domain.RecommendEntry{
		{URL: "https://example.com/1", Title: "Test 1", PostedAt: time.Now()},
		{URL: "https://example.com/2", Title: "Test 2", PostedAt: time.Now()},
	}

	for _, entry := range testEntries {
		cache.entries = append(cache.entries, entry)
		cache.urlSet[cache.normalizeURL(entry.URL)] = true
	}

	err := cache.saveToFile()
	if err != nil {
		t.Fatalf("saveToFile failed: %v", err)
	}

	// ファイルの内容を確認
	data, err := os.ReadFile(config.FilePath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	// JSON形式の確認
	for i, line := range lines {
		var entry domain.RecommendEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

func TestFileRecommendCache_acquireAndReleaseLock(t *testing.T) {
	tmpDir := t.TempDir()

	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "cache.jsonl"),
		MaxEntries:    100,
		RetentionDays: 7,
	}

	cache1 := NewFileRecommendCache(config)

	t.Run("ロック取得成功", func(t *testing.T) {
		err := cache1.acquireLock()
		if err != nil {
			t.Fatalf("acquireLock failed: %v", err)
		}

		// ロックファイルが存在することを確認
		if _, err := os.Stat(cache1.lockPath); os.IsNotExist(err) {
			t.Error("Lock file should exist")
		}
	})

	t.Run("同じロックの二重取得エラー", func(t *testing.T) {
		cache2 := NewFileRecommendCache(config)
		err := cache2.acquireLock()
		if err == nil {
			t.Error("Second lock acquisition should fail")
		}
		if !strings.Contains(err.Error(), "lock file exists") {
			t.Errorf("Expected lock conflict error, got: %v", err)
		}
	})

	t.Run("ロック解放成功", func(t *testing.T) {
		err := cache1.releaseLock()
		if err != nil {
			t.Fatalf("releaseLock failed: %v", err)
		}

		// ロックファイルが削除されていることを確認
		if _, err := os.Stat(cache1.lockPath); !os.IsNotExist(err) {
			t.Error("Lock file should be removed")
		}
	})
}

func TestFileRecommendCache_cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "cleanup_test.jsonl"),
		MaxEntries:    100,
		RetentionDays: 1,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()
	defer cache.Close()

	// テストデータ: 古いエントリと新しいエントリ
	oldEntry := domain.RecommendEntry{
		URL:      "https://example.com/old",
		Title:    "Old Article",
		PostedAt: time.Now().AddDate(0, 0, -2), // 2日前
	}
	newEntry := domain.RecommendEntry{
		URL:      "https://example.com/new",
		Title:    "New Article",
		PostedAt: time.Now(),
	}

	cache.entries = []domain.RecommendEntry{oldEntry, newEntry}
	cache.urlSet[cache.normalizeURL(oldEntry.URL)] = true
	cache.urlSet[cache.normalizeURL(newEntry.URL)] = true

	// クリーンアップ実行
	cache.cleanup()

	// 古いエントリが削除され、新しいエントリが残っていることを確認
	if len(cache.entries) != 1 {
		t.Errorf("Expected 1 entry after cleanup, got %d", len(cache.entries))
	}
	if cache.entries[0].URL != newEntry.URL {
		t.Error("Wrong entry remained after cleanup")
	}
	if len(cache.urlSet) != 1 {
		t.Errorf("Expected 1 URL in urlSet after cleanup, got %d", len(cache.urlSet))
	}
	if !cache.urlSet[cache.normalizeURL(newEntry.URL)] {
		t.Error("New entry URL should remain in urlSet")
	}
	if cache.urlSet[cache.normalizeURL(oldEntry.URL)] {
		t.Error("Old entry URL should be removed from urlSet")
	}
}

func TestFileRecommendCache_cleanupByMaxEntries(t *testing.T) {
	tmpDir := t.TempDir()
	config := &entity.CacheConfig{
		Enabled:       true,
		FilePath:      filepath.Join(tmpDir, "max_entries_test.jsonl"),
		MaxEntries:    2,
		RetentionDays: 30,
	}

	cache := NewFileRecommendCache(config)
	cache.Initialize()
	defer cache.Close()

	// 最大エントリ数を超えるテストデータ
	entries := []domain.RecommendEntry{
		{URL: "https://example.com/1", Title: "Article 1", PostedAt: time.Now().AddDate(0, 0, -3)},
		{URL: "https://example.com/2", Title: "Article 2", PostedAt: time.Now().AddDate(0, 0, -2)},
		{URL: "https://example.com/3", Title: "Article 3", PostedAt: time.Now().AddDate(0, 0, -1)},
	}

	for _, entry := range entries {
		cache.entries = append(cache.entries, entry)
		cache.urlSet[cache.normalizeURL(entry.URL)] = true
	}

	// クリーンアップ実行
	cache.cleanupByMaxEntries()

	// 最大エントリ数まで削減されていることを確認
	if len(cache.entries) != config.MaxEntries {
		t.Errorf("Expected %d entries after cleanup, got %d", config.MaxEntries, len(cache.entries))
	}
	if len(cache.urlSet) != config.MaxEntries {
		t.Errorf("Expected %d URLs in urlSet after cleanup, got %d", config.MaxEntries, len(cache.urlSet))
	}

	// 古いエントリ（Article 1）が削除されていることを確認
	for _, entry := range cache.entries {
		if entry.URL == "https://example.com/1" {
			t.Error("Oldest entry should be removed")
		}
	}
	if cache.urlSet[cache.normalizeURL("https://example.com/1")] {
		t.Error("Oldest entry URL should be removed from urlSet")
	}
}

func TestFileRecommendCache_normalizeURL(t *testing.T) {
	cache := &FileRecommendCache{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"https://example.com/", "https://example.com"},
		{"https://example.com", "https://example.com"},
		{"https://example.com/path/", "https://example.com/path"},
		{"https://example.com/path", "https://example.com/path"},
	}

	for _, tc := range testCases {
		result := cache.normalizeURL(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeURL(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestFileRecommendCache_ErrorHandling(t *testing.T) {
	t.Run("不正なディレクトリでの初期化エラー", func(t *testing.T) {
		// 読み取り専用ディレクトリを作成
		tmpDir := t.TempDir()
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		os.Mkdir(readOnlyDir, 0444)
		defer os.Chmod(readOnlyDir, 0755) // クリーンアップのために権限を戻す

		config := &entity.CacheConfig{
			Enabled:       true,
			FilePath:      filepath.Join(readOnlyDir, "cache.jsonl"),
			MaxEntries:    100,
			RetentionDays: 7,
		}

		cache := NewFileRecommendCache(config)
		err := cache.Initialize()
		if err == nil {
			cache.Close()
			t.Error("Initialize should fail with readonly directory")
		}
	})

	t.Run("壊れたJSONファイルの読み込み", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheFile := filepath.Join(tmpDir, "corrupted.jsonl")
		config := &entity.CacheConfig{
			Enabled:       true,
			FilePath:      cacheFile,
			MaxEntries:    100,
			RetentionDays: 7,
		}

		// 壊れたJSONファイルを作成（現在時刻にして削除されないようにする）
		validEntry := domain.RecommendEntry{
			URL:      "https://example.com/valid",
			Title:    "valid",
			PostedAt: time.Now(),
		}
		validJSON, _ := json.Marshal(validEntry)
		fileContent := "invalid json line\n" + string(validJSON) + "\n"
		os.WriteFile(cacheFile, []byte(fileContent), 0644)

		cache := NewFileRecommendCache(config)
		err := cache.Initialize()
		if err != nil {
			t.Fatalf("Initialize should handle corrupted JSON gracefully: %v", err)
		}

		// 有効なエントリのみが読み込まれていることを確認
		if len(cache.entries) != 1 {
			t.Errorf("Expected 1 valid entry, got %d. Entries: %+v", len(cache.entries), cache.entries)
		}
		if len(cache.urlSet) != 1 {
			t.Errorf("Expected 1 URL in urlSet, got %d", len(cache.urlSet))
		}

		cache.Close()
	})
}
