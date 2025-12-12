package testutil

// BoolPtr はbool値へのポインタを返すテスト用ヘルパー関数
func BoolPtr(b bool) *bool {
	return &b
}

// StringPtr は文字列値へのポインタを返すテスト用ヘルパー関数
func StringPtr(s string) *string {
	return &s
}
