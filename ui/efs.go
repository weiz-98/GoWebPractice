package ui

import (
	"embed"
)

//go:embed "html" "static"
var Files embed.FS

// 嵌入式檔案系統始終根位於包含 go:embed 指令的目錄中。
// 我們的 Files 變數包含一個 embed.FS 嵌入式檔案系統，該檔案系統的根目錄是我們的 ui 目錄。
