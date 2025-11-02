package infra

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

var levelColorMap = map[slog.Level]*color.Color{
	slog.LevelDebug: color.New(color.FgHiBlack), // 灰色
	slog.LevelInfo:  color.New(color.FgGreen),   // 緑色
	slog.LevelWarn:  color.New(color.FgYellow),  // 黄色
	slog.LevelError: color.New(color.FgRed),     // 赤色
}

// SimpleHandler は時刻 ログレベル ログメッセージの形式で出力するカスタムハンドラー
type SimpleHandler struct {
	opts slog.HandlerOptions
	mu   sync.Mutex
	out  io.Writer
}

// formatAttr はslog.Valueを文字列にフォーマットする
// GroupValueの場合は展開して表示する
func formatAttr(key string, value slog.Value) string {
	switch value.Kind() {
	case slog.KindGroup:
		// グループの場合は各属性を展開
		attrs := value.Group()
		if len(attrs) == 0 {
			return fmt.Sprintf("%s={}", key)
		}
		parts := make([]string, 0, len(attrs))
		for _, attr := range attrs {
			parts = append(parts, formatAttr(attr.Key, attr.Value.Resolve()))
		}
		return fmt.Sprintf("%s={%s}", key, strings.Join(parts, " "))
	default:
		return fmt.Sprintf("%s=%v", key, value.Any())
	}
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

	// ログレベルを取得し、色を適用
	levelStr := r.Level.String()
	if c, ok := levelColorMap[r.Level]; ok {
		level = c.Sprint(levelStr)
	} else {
		level = levelStr
	}

	// メッセージを取得
	msg = r.Message

	// 属性を追加
	attrs := make([]string, 0)
	r.Attrs(func(a slog.Attr) bool {
		// Resolve()を使ってLogValue()メソッドを呼び出す
		value := a.Value.Resolve()
		attrs = append(attrs, formatAttr(a.Key, value))
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
// verboseがtrueの場合はDEBUGレベルをstderrに出力、falseの場合はログを出力しない
func InitLogger(verbose bool) {
	var handler slog.Handler
	if !verbose {
		// -vなしの場合はログを出力しない
		handler = slog.NewTextHandler(io.Discard, nil)
	} else {
		// -vありの場合はDEBUGレベル以上をstderrに出力
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		handler = NewSimpleHandler(os.Stderr, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
