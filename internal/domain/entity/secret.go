package entity

import "log/slog"

// SecretString は機密情報を扱うための型
// ログ出力や文字列化の際に自動的にマスクされる
type SecretString struct {
	value string
}

// Value は元の値を返す
func (s SecretString) Value() string {
	return s.value
}

// String は fmt.Stringer インターフェースを実装し、マスクされた文字列を返す
func (s SecretString) String() string {
	return "[REDACTED]"
}

// IsEmpty は値が空かどうかを判定する
func (s SecretString) IsEmpty() bool {
	return s.value == ""
}

// LogValue は slog.LogValuer インターフェースを実装し、ログ出力時にマスクされた値を返す
func (s SecretString) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

// UnmarshalText は encoding.TextUnmarshaler インターフェースを実装し、バイト列から値を設定する
func (s *SecretString) UnmarshalText(text []byte) error {
	s.value = string(text)
	return nil
}

// MarshalText は encoding.TextMarshaler インターフェースを実装し、マスクされた値をバイト列として返す
func (s SecretString) MarshalText() ([]byte, error) {
	return []byte("[REDACTED]"), nil
}
