//go:build e2e

package mock

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// MockMisskeyReceiver はMisskeyのノート作成リクエストを受信・記録するモックサーバー
type MockMisskeyReceiver struct {
	mu    sync.RWMutex
	notes []string
}

// NewMockMisskeyReceiver はMockMisskeyReceiverの新しいインスタンスを生成する
func NewMockMisskeyReceiver() *MockMisskeyReceiver {
	return &MockMisskeyReceiver{
		notes: make([]string, 0),
	}
}

// ServeHTTP はhttp.Handlerインターフェースを実装し、API受信を処理する
func (m *MockMisskeyReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ受け付ける
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// リクエストボディを読み取る
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// JSONをパースしてノートを抽出
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// ノートのテキストを記録
	if text, ok := payload["text"].(string); ok {
		m.mu.Lock()
		m.notes = append(m.notes, text)
		m.mu.Unlock()
	}

	// Misskeyは作成されたノート情報を返す（簡易的なレスポンス）
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"id":        "mock-note-id",
		"createdAt": "2024-01-01T00:00:00.000Z",
	}
	// レスポンスの書き込みエラーは通常発生しないが、
	// クライアントが接続を切断した場合などに備えてエラーを無視
	_ = json.NewEncoder(w).Encode(response)
}

// ReceivedNote はノートが少なくとも1つ受信されたかを返す
func (m *MockMisskeyReceiver) ReceivedNote() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.notes) > 0
}

// GetNotes は受信したノートの一覧を返す
func (m *MockMisskeyReceiver) GetNotes() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// スライスのコピーを返す
	result := make([]string, len(m.notes))
	copy(result, m.notes)
	return result
}

// GetLastNote は最後に受信したノートを返す
// ノートが一つもない場合は空文字列を返す
func (m *MockMisskeyReceiver) GetLastNote() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.notes) == 0 {
		return ""
	}
	return m.notes[len(m.notes)-1]
}

// Reset は受信したノートをすべてクリアする
func (m *MockMisskeyReceiver) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notes = make([]string, 0)
}
