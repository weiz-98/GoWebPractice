package assert

import (
	"strings"
	"testing"
)

func Equal[T comparable](t *testing.T, actual, expected T) {
	t.Helper()
	// Helper 讓 Go 測試運行器將在輸出中報告呼叫 Equal() 函數的程式碼的檔案名稱和行號
	if actual != expected {
		t.Errorf("got: %v; want: %v", actual, expected)
	}
}

func StringContains(t *testing.T, actual, expectedSubstring string) {
	t.Helper()
	if !strings.Contains(actual, expectedSubstring) {
		t.Errorf("got: %q; expected to contain: %q", actual, expectedSubstring)
	}
}
