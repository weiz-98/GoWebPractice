package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// Change the signature of the home handler so it is defined as a method against
// *application.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w) // Use the notFound() helper
		return
	}
	// Initialize a slice containing the paths to the two files. It's important
	// to note that the file containing our base template must be the *first* // file in the slice.
	files := []string{
		"./ui/html/base.tmpl",
		"./ui/html/partials/nav.tmpl",
		"./ui/html/pages/home.tmpl"}
	// 使用 template 包的 ParseFiles 函數來讀取和解析上述提到的模板文件
	ts, err := template.ParseFiles(files...)
	if err != nil {
		// Because the home handler function is now a method against application
		// it can access its fields, including the error logger. We'll write the log // message to this instead of the standard logger.
		app.serverError(w, err) // Use the serverError() helper.
		http.Error(w, "Internal Server Error", 500)
		return
	}
	// 使用 ExecuteTemplate 方法將名為 base 的模板輸出到 HTTP response body.
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, err) // Use the serverError() helper.
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// Extract the value of the id parameter from the query string and try to
	// convert it to an integer using the strconv.Atoi() function. If it can't
	// be converted to an integer, or the value is less than 1, we return a 404 page
	// not found response.
	// strconv.Atoi 確保了程序能正確處理非預期的輸入，並在出現錯誤時提供適當的回應。
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 1 {
		app.notFound(w) // Use the notFound() helper.
		return
	}
	// Use the fmt.Fprintf() function to interpolate the id value with our response
	// and write it to the http.ResponseWriter.
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	// Use r.Method to check whether the request is using POST or not.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		// Use the http.Error() function to send a 405 status code and string as the response body.
		// 最大的區別是我們現在將 http.ResponseWriter 傳遞給另一個函數，該函數會為我們向使用者發送回應。
		// 使用 net/http 套件中的常數來表示 HTTP 方法和狀態代碼，而不是自己編寫字串和整數。
		// http.Error(w, "This Method Not Allowed", 405)
		app.clientError(w, http.StatusMethodNotAllowed) // Use the clientError() helper.
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"name":"Ian"}`))

}
