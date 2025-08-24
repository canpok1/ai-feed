package domain

// ReleaseInfo はリリース情報を表す構造体
type ReleaseInfo struct {
	// Version はリリースのバージョン番号
	Version string
	// AssetURL はバイナリのダウンロードURL
	AssetURL string
	// ReleaseNotes はリリースノート
	ReleaseNotes string
}

// Updater はバイナリ更新機能のインターフェース
type Updater interface {
	// GetCurrentVersion は現在のバージョンを取得する
	GetCurrentVersion() (string, error)
	// GetLatestVersion は最新のリリース情報を取得する
	GetLatestVersion() (*ReleaseInfo, error)
	// UpdateBinary は指定されたバージョンにバイナリを更新する
	UpdateBinary(version string) error
}
