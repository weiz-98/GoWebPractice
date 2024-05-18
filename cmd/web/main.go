package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	// Define a new command-line flag with the name 'addr', a default value of ":4000"
	// and some short help text explaining what the flag controls. The value of the
	// flag will be stored in the addr variable at runtime.
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Importantly, we use the flag.Parse() function to parse the command-line flag.
	// This reads in the command-line flag value and assigns it to the addr
	// variable. You need to call this *before* you use the addr variable
	// otherwise it will always contain the default value of ":4000". If any errors are
	// encountered during parsing the application will be terminated.
	flag.Parse()
	// Use the http.NewServeMux() function to initialize a new servemux, then
	// register the home function as the handler for the "/" URL pattern.
	mux := http.NewServeMux()

	// Create a file server which serves files out of the "./ui/static" directory.
	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	fileServer := http.FileServer(http.Dir("./ui/static/"))
	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/". For matching paths, we strip the
	// "/static" prefix before the request reaches the file server.
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// 但這個home功能只是普通功能； 它沒有 ServeHTTP() 方法。 所以它本身並不是一個處理程序。
	// 相反，我們可以使用 http.HandlerFunc() handler 將其轉換為處理程序
	// http.HandlerFunc() 適配器的工作原理是自動將 ServeHTTP() 方法新增至 home 函數。
	mux.Handle("/", http.HandlerFunc(home))

	mux.HandleFunc("/snippet/view", snippetView)
	mux.HandleFunc("/snippet/create", snippetCreate)
	log.Printf("Starting server on %s", *addr)
	// Use the http.ListenAndServe() function to start a new web server. We pass in
	// two parameters: the TCP network address to listen on (in this case ":4000")
	// and the servermux we just created.

	// The value returned from the flag.String() function is a pointer to the flag
	// value, not the value itself. So we need to dereference the pointer (i.e.
	// prefix it with the * symbol) before using it. Note that we're using the
	// log.Printf() function to interpolate the address with the log message.
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}

// 當我們的伺服器收到一個新的 HTTP 請求時，它會呼叫 servemux 的 ServeHTTP() 方法。
// 這會根據請求 URL 路徑尋找相關處理程序，然後呼叫該處理程序的 ServeHTTP() 方法。
// 您可以將 Go Web 應用程式視為一系列依序呼叫的 ServeHTTP() 方法。

// 所有傳入的 HTTP 請求都在它們自己的 goroutine 中提供服務。
// 對於繁忙的伺服器，這意味著處理程序中的程式碼或由處理程序呼叫的程式碼很可能會同時執行
