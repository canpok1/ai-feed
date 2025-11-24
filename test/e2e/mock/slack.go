//go:build e2e

package mock

import (
	"encoding/json"
	"io"
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

// ServeHTTP はhttp.Handlerインターフェースを実装し、Webhook受信を処理する
func (m *MockSlackReceiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	// JSONをパースしてメッセージを抽出
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// メッセージを記録
	if text, ok := payload["text"].(string); ok {
		m.mu.Lock()
		m.messages = append(m.messages, text)
		m.mu.Unlock()
	}

	// Slackは通常 "ok" を返す
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		// テスト用モックサーバーなので、エラーは無視するがログ出力のみ行う
		// 実際のテストではクライアントの接続切断などでエラーが発生する可能性がある
		_ = err // エラーは無視
	}
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
