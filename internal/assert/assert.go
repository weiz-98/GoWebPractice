package assert

import (
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	// Helper 讓 Go 測試運行器將在輸出中報告呼叫 Equal() 函數的程式碼的檔案名稱和行號
	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}
