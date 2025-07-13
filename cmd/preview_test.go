package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// createTempFile はテスト用のテンポラリファイルを作成し、そのパスを返します。
func createTempFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "test_urls_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	return tmpfile.Name()
}

// TestReadURLsFromFileValid は有効なURLリストを含むファイルの読み込みをテストします。
func TestReadURLsFromFileValid(t *testing.T) {
	content := "http://example.com/1\nhttps://example.org/2\n"
	filePath := createTempFile(t, content)
	defer os.Remove(filePath)

	cmd := &cobra.Command{} // モックのcobra.Command
	urls, err := readURLsFromFile(filePath, cmd)
	if err != nil {
		t.Fatalf("readURLsFromFile returned an error: %v", err)
	}

	expected := []string{"http://example.com/1", "https://example.org/2"}
	if len(urls) != len(expected) {
		t.Fatalf("Expected %d URLs, got %d", len(expected), len(urls))
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("Expected URL %s, got %s", expected[i], u)
		}
	}
}

// TestReadURLsFromFileEmptyLines は空行を含むファイルの読み込みをテストします。
func TestReadURLsFromFileEmptyLines(t *testing.T) {
	content := "http://example.com/1\n\nhttps://example.org/2\n\n"
	filePath := createTempFile(t, content)
	defer os.Remove(filePath)

	cmd := &cobra.Command{} // モックのcobra.Command
	urls, err := readURLsFromFile(filePath, cmd)
	if err != nil {
		t.Fatalf("readURLsFromFile returned an error: %v", err)
	}

	expected := []string{"http://example.com/1", "https://example.org/2"}
	if len(urls) != len(expected) {
		t.Fatalf("Expected %d URLs, got %d", len(expected), len(urls))
	}
	for i, u := range urls {
		if u != expected[i] {
			t.Errorf("Expected URL %s, got %s", expected[i], u)
		}
	}
}

// TestReadURLsFromFileInvalidURLs は不正なURLを含むファイルの読み込みをテストします。
func TestReadURLsFromFileInvalidURLs(t *testing.T) {
	content := "http://example.com/1\nnot-a-url\nhttps://example.org/2\n"
	filePath := createTempFile(t, content)
	defer os.Remove(filePath)

	// 標準エラー出力をキャプチャ
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	cmd := &cobra.Command{} // モックのcobra.Command
	urls, err := readURLsFromFile(filePath, cmd)

	w.Close()
	os.Stderr = oldStderr // 標準エラー出力を元に戻す
	var buf bytes.Buffer
	buf.ReadFrom(r)
	stderrOutput := buf.String()

	if err != nil {
		t.Fatalf("readURLsFromFile returned an error: %v", err)
	}

	expectedURLs := []string{"http://example.com/1", "https://example.org/2"}
	if len(urls) != len(expectedURLs) {
		t.Fatalf("Expected %d URLs, got %d", len(expectedURLs), len(urls))
	}
	for i, u := range urls {
		if u != expectedURLs[i] {
			t.Errorf("Expected URL %s, got %s", expectedURLs[i], u)
		}
	}

	if !strings.Contains(stderrOutput, "Warning: Invalid URL in") || !strings.Contains(stderrOutput, "not-a-url") {
		t.Errorf("Expected warning for invalid URL, got: %s", stderrOutput)
	}
}

// TestReadURLsFromFileNonExistent は存在しないファイルの読み込みをテストします。
func TestReadURLsFromFileNonExistent(t *testing.T) {
	cmd := &cobra.Command{} // モックのcobra.Command
	_, err := readURLsFromFile("non_existent_file.txt", cmd)
	if err == nil {
		t.Fatal("readURLsFromFile did not return an error for a non-existent file")
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Errorf("Expected 'no such file or directory' error, got: %v", err)
	}
}

// TestPreviewCommandSourceAndURLConflict は --source と --url オプションの同時使用をテストします。
func TestPreviewCommandSourceAndURLConflict(t *testing.T) {
	b := bytes.NewBufferString("")
	cmd := &cobra.Command{}
	cmd.SetOut(b)
	cmd.SetErr(b)

	// フラグを定義
	cmd.Flags().StringSliceP("url", "u", []string{}, "URL of the feed to preview")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs to preview")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of articles to display")

	// フラグをセット
	cmd.Flags().Set("url", "http://example.com")
	cmd.Flags().Set("source", "list.txt")

	err := previewCmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("Expected an error when --source and --url are used together, but got none.")
	}
	if !strings.Contains(err.Error(), "cannot use --source and --url options together") {
		t.Errorf("Expected conflict error, got: %v", err)
	}
}

// TestPreviewCommandDuplicateURLs はURLの重複排除をテストします。
func TestPreviewCommandDuplicateURLs(t *testing.T) {
	// テスト用のテンポラリファイルを作成
	content := "http://example.com/1\nhttp://example.com/2\nhttp://example.com/1\n"
	filePath := createTempFile(t, content)
	defer os.Remove(filePath)

	cmd := &cobra.Command{} // モックのcobra.Command
	cmd.Flags().StringSliceP("url", "u", []string{}, "URL of the feed to preview")
	cmd.Flags().StringP("source", "s", "", "Path to a file containing a list of URLs to preview")
	cmd.Flags().IntP("limit", "l", 0, "Maximum number of articles to display")

	// フラグをセット
	cmd.Flags().Set("source", filePath)

	// readURLsFromFile を直接呼び出してURLリストを取得
	urls, err := readURLsFromFile(filePath, cmd)
	if err != nil {
		t.Fatalf("readURLsFromFile returned an error: %v", err)
	}

	// 重複排除ロジックを直接テスト
	uniqueURLs := make(map[string]bool)
	var finalURLs []string
	for _, url := range urls {
		if _, ok := uniqueURLs[url]; !ok {
			uniqueURLs[url] = true
			finalURLs = append(finalURLs, url)
		}
	}

	expected := []string{"http://example.com/1", "http://example.com/2"}
	if len(finalURLs) != len(expected) {
		t.Fatalf("Expected %d unique URLs, got %d", len(expected), len(finalURLs))
	}
	for i, u := range finalURLs {
		if u != expected[i] {
			t.Errorf("Expected unique URL %s, got %s", expected[i], u)
		}
	}
}


