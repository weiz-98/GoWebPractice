package main

import (
	"crypto/tls"
	"database/sql" // New import
	"flag"
	"html/template"
	"log"
	"net/http"
	"os" // New import
	"time"

	//you can find it at the top of the go.mod file.
	"GoWebPractice/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form"
	_ "github.com/go-sql-driver/mysql" // 表示這個包會被導入，但不會在程式碼中直接使用這個包中的任何函數或類型。這種導入方式通常用於其副作用（side effects）。
)

// Add a snippets field to the application struct. This will allow us to
// make the SnippetModel object available to our handlers.
// 使我們的模型成為一個單一的、整齊封裝的對象，我們可以輕鬆地初始化該對象，然後將其作為依賴項傳遞給我們的處理程序
type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	// Define a new command-line flag for the MySQL DSN string.
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// To keep the main() function tidy I've put the code for creating a connection
	// pool into the separate openDB() function below. We pass openDB() the DSN
	// from the command-line flag.
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}

	// We also defer a call to db.Close(), so that the connection pool is closed
	// before the main() function exits.
	defer db.Close()

	// Initialize a new template cache...
	templateCache, err := newTemplateCache()
	if err != nil {
		errorLog.Fatal(err)
	}
	// Initialize a decoder instance...
	formDecoder := form.NewDecoder()

	// Use the scs.New() function to initialize a new session manager. Then we
	// configure it to use our MySQL database as the session store, and set a
	// lifetime of 12 hours (so that sessions automatically expire 12 hours
	// after first being created).
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	// Make sure that the Secure attribute is set on our session cookies.
	// Setting this means that the cookie will only be sent by a user's web
	// browser when a HTTPS connection is being used (and won't be sent over an
	// unsecure HTTP connection).
	sessionManager.Cookie.Secure = true

	app := &application{
		errorLog:       errorLog,
		infoLog:        infoLog,
		snippets:       &models.SnippetModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
	stack := app.createStack( //透過建立 createStack 把所有 middleware 串接起來
		app.logRequest,
		app.secureHeaders,
		app.recoverPanic,
		app.sessionManager.LoadAndSave,
	)
	// Initialize a tls.Config struct to hold the non-default TLS settings we
	// want the server to use. In this case the only thing that we're changing
	// is the curve preferences value, so that only elliptic curves with
	// assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	// Set the server's TLSConfig field to use the tlsConfig variable we just
	// created.
	srv := &http.Server{ // 使用指針型態才可以在整個專案共享服務器配置
		Addr:     *addr,
		ErrorLog: errorLog,
		// Call the new app.routes() method to get the servemux containing our routes.
		Handler:   stack(app.routes()), // 1.22 版本之後可以直接 wrap 在 router 之外
		TLSConfig: tlsConfig,
		// Add Idle, Read and Write timeouts to the server.
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 524288,
	}

	infoLog.Printf("Starting server on %s", *addr)
	// Use the ListenAndServeTLS() method to start the HTTPS server. We
	// pass in the paths to the TLS certificate and corresponding private key as
	// the two parameters.
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

// The openDB() function wraps sql.Open() and returns a sql.DB connection pool
// for a given DSN.
func openDB(dsn string) (*sql.DB, error) {
	// sql.Open() 函數實際上並沒有創建任何連接，它所做的只是初始化池以供將來使用
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// 與資料庫的實際連線是在第一次需要時延遲建立的。
	// 因此，為了驗證一切設定是否正確，我們需要使用 db.Ping() 方法來建立連線並檢查是否有任何錯誤。
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}
