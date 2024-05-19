package main

import (
	"database/sql" // New import
	"flag"
	"html/template"
	"log"
	"net/http"
	"os" // New import

	//you can find it at the top of the go.mod file.
	"GoWebPractice/internal/models"

	_ "github.com/go-sql-driver/mysql" // 表示這個包會被導入，但不會在程式碼中直接使用這個包中的任何函數或類型。這種導入方式通常用於其副作用（side effects）。
)

// Add a snippets field to the application struct. This will allow us to
// make the SnippetModel object available to our handlers.
// 使我們的模型成為一個單一的、整齊封裝的對象，我們可以輕鬆地初始化該對象，然後將其作為依賴項傳遞給我們的處理程序
type application struct {
	errorLog      *log.Logger
	infoLog       *log.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
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

	// Initialize a models.SnippetModel instance and add it to the application
	// dependencies.
	app := &application{
		errorLog:      errorLog,
		infoLog:       infoLog,
		snippets:      &models.SnippetModel{DB: db},
		templateCache: templateCache,
	}
	stack := app.createStack( //透過建立 createStack 把所有 middleware 串接起來
		app.logRequest,
		app.secureHeaders,
		app.recoverPanic,
	)
	srv := &http.Server{ // 使用指針型態才可以在整個專案共享服務器配置
		Addr:     *addr,
		ErrorLog: errorLog,
		// Call the new app.routes() method to get the servemux containing our routes.
		Handler: stack(app.routes()), // 1.22 版本之後可以直接 wrap 在 router 之外
	}

	infoLog.Printf("Starting server on %s", *addr)
	// Because the err variable is now already declared in the code above, we need
	// to use the assignment operator = here, instead of the := 'declare and assign'
	// operator.
	err = srv.ListenAndServe()
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
