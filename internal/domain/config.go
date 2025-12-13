package domain

// ConfigInitRepository は設定ファイルの初期化を担当するインターフェース
type ConfigInitRepository interface {
	// SaveWithTemplate はテンプレートを使用して設定ファイルを保存する
	SaveWithTemplate() error
}
