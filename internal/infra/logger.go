package infra

import (
	"log/slog"
	"os"
	"time"
)

// InitLogger はslogを使用したロガーを初期化する
// verboseがtrueの場合はDEBUGレベル、falseの場合はINFOレベルで設定する
func InitLogger(verbose bool) {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// タイムスタンプをRFC3339形式に変更
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   a.Key,
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
