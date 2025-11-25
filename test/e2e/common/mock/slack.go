//go:build e2e

package mock

import (
	"encoding/json"
	"net/http"
	"sync"
)

// MockSlackReceiver はSlack Webhookのメッセージを受信・記録するモックサーバー
type MockSlackReceiver struct {
	mu       sync.RWMutex
	messages []string
}

// NewMockSlackReceiver はMockSlackReceiverの新しいインスタンスを生成する
func NewMockSlackReceiver() *MockSlackReceiver {
	return &MockSlackReceiver{
		messages: make([]string, 0),
	}
}

// ServeHTTP はhttp.Handlerインターフェースを実装し、Slack API (chat.postMessage) を模倣する
func (m *MockSlackReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// POSTメソッドのみ受け付ける
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Content-Typeに応じてパース方法を変える
	contentType := r.Header.Get("Content-Type")
	var text string

	if contentType == "application/json" {
		// JSON形式のリクエストを処理
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if t, ok := payload["text"].(string); ok {
			text = t
		}
	} else {
		// application/x-www-form-urlencoded形式のリクエストを処理
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		text = r.FormValue("text")
	}

	// メッセージを記録
	if text != "" {
		m.mu.Lock()
		m.messages = append(m.messages, text)
		m.mu.Unlock()
	}

	// Slack API形式のレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"ok":      true,
		"channel": "C1234567890",
		"ts":      "1234567890.123456",
	}
	// レスポンスの書き込みエラーは通常発生しないが、
	// クライアントが接続を切断した場合などに備えてエラーを無視
	_ = json.NewEncoder(w).Encode(response)
}

// ReceivedMessage はメッセージが少なくとも1つ受信されたかを返す
func (m *MockSlackReceiver) ReceivedMessage() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages) > 0
}

// GetMessages は受信したメッセージの一覧を返す
func (m *MockSlackReceiver) GetMessages() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	// スライスのコピーを返す
	result := make([]string, len(m.messages))
	copy(result, m.messages)
	return result
}

// GetLastMessage は最後に受信したメッセージを返す
// メッセージが一つもない場合は空文字列を返す
func (m *MockSlackReceiver) GetLastMessage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if len(m.messages) == 0 {
		return ""
	}
	return m.messages[len(m.messages)-1]
}

// Reset は受信したメッセージをすべてクリアする
func (m *MockSlackReceiver) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = make([]string, 0)
}
