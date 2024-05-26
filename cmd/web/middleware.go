package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/justinas/nosurf" // 使用雙重提交 cookie 模式來防止攻擊。
	// 在此模式中，會產生隨機 CSRF 令牌並將其透過 CSRF cookie 傳送給使用者。
	// 然後，將此 CSRF 令牌新增至每個容易受到 CSRF 攻擊的 HTML 表單中的隱藏欄位中。
	// 當提交表單時，兩個套件都會使用一些中間件來檢查隱藏欄位值和 cookie 值是否符合
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		next.ServeHTTP(w, r)
	})
}

type warpWriter struct {
	http.ResponseWriter
	statusCode int
}

// The method writeHeader is renamed to WriteHeader to correctly override the ResponseWriter interface's WriteHeader method.
func (w *warpWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &warpWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(wrapped, r)
		app.infoLog.Printf("%d %s %s %s", wrapped.statusCode, r.Method, r.URL.RequestURI(), time.Since(start))
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event
		// of a panic as Go unwinds the stack).
		defer func() {
			// Use the builtin recover function to check if there has been a
			// panic or not. If there has...
			if err := recover(); err != nil {
				// Set a "Connection: close" header on the response.
				w.Header().Set("Connection", "close")
				// Call the app.serverError helper method to return a 500
				// Internal Server response.
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		if !app.isAuthenticated(r) {
			// Add the path that the user is trying to access to their session
			// data.
			app.sessionManager.Put(r.Context(), "redirectPathAfterLogin", r.URL.Path)
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")
		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Create a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly attributes set.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true, Path: "/", Secure: true,
	})
	return csrfHandler
}

// isAuthenticated() 幫助器可能在每個請求週期中被多次呼叫。 目前我們使用它兩次——
// 一次在 requireAuthentication() 中間件中，另一次在 newTemplateData() 幫助器中。
// 因此，如果我們直接從 isAuthenticated() 幫助程式查詢資料庫，我們最終會在每個請求期間對資料庫進行重複的往返
// 解決方式是做成 middleware 減少 資料庫 query 次數
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the authenticatedUserID value from the session using the
		// GetInt() method. This will return the zero value for an int (0) if no
		// "authenticatedUserID" value is in the session -- in which case we
		// call the next handler in the chain as normal and return.
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 {
			next.ServeHTTP(w, r)
			return
		}
		// Otherwise, we check to see if a user with that ID exists in our
		// database.
		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		// If a matching user is found, we know we know that the request is
		// coming from an authenticated user who exists in our database. We
		// create a new copy of the request (with an isAuthenticatedContextKey
		// value of true in the request context) and assign it to r.
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

//authMiddleware：通過 r.WithContext(ctx) 創建了一個新的請求對象，該對象包含了更新的上下文，
// 而不是直接修改原始請求。這樣做可以保證原始請求對象的不可變性。
//handler：可以安全地從上下文中讀取數據，而不用擔心其他中間件或處理器對原始請求的修改。
