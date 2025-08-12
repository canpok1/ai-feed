package infra

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

// SimpleHandler は時刻 ログレベル ログメッセージの形式で出力するカスタムハンドラー
type SimpleHandler struct {
	opts slog.HandlerOptions
	mu   sync.Mutex
	out  io.Writer
}

// NewSimpleHandler は新しいSimpleHandlerを作成する
func NewSimpleHandler(out io.Writer, opts *slog.HandlerOptions) *SimpleHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &SimpleHandler{
		opts: *opts,
		out:  out,
	}
}

// Enabled はログレベルが有効かどうかを判定する
func (h *SimpleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

// Handle はログレコードを処理して出力する
func (h *SimpleHandler) Handle(_ context.Context, r slog.Record) error {
	var timestamp string
	var level string
	var msg string

	// タイムスタンプを取得
	timestamp = r.Time.Format(time.RFC3339)

	// ログレベルを取得
	level = r.Level.String()

	// メッセージを取得
	msg = r.Message

	// 属性を追加
	attrs := make([]string, 0)
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		return true
	})

	h.mu.Lock()
	defer h.mu.Unlock()

	// 出力形式: 時刻 ログレベル ログメッセージ [属性...]
	if len(attrs) > 0 {
		_, err := fmt.Fprintf(h.out, "%s %s %s", timestamp, level, msg)
		if err != nil {
			return err
		}
		for _, attr := range attrs {
			_, err = fmt.Fprintf(h.out, " %s", attr)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(h.out)
		return err
	}
	_, err := fmt.Fprintf(h.out, "%s %s %s\n", timestamp, level, msg)
	return err
}

// WithAttrs は属性を追加したハンドラーを返す（この実装では新しいハンドラーを返す）
func (h *SimpleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup はグループを追加したハンドラーを返す（この実装では新しいハンドラーを返す）
func (h *SimpleHandler) WithGroup(name string) slog.Handler {
	return h
}

// InitLogger はslogを使用したロガーを初期化する
// verboseがtrueの場合はDEBUGレベル、falseの場合はINFOレベルで設定する
func InitLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := NewSimpleHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
