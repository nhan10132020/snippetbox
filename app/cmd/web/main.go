package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/joho/godotenv"
	"github.com/nhan10132020/snippetbox/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type application struct {
	errorLog       *log.Logger
	infoLog        *log.Logger
	snippets       *models.SnippetModel
	users          *models.UserModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")

	db_host := os.Getenv("DB_HOST")
	db_name := os.Getenv("DB_DATABASE_NAME")
	db_username := os.Getenv("DB_USERNAME")
	db_password := os.Getenv("DB_PASSWORD")
	db_port := os.Getenv("DB_PORT")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", db_username, db_password, db_host, db_port, db_name)

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, sqlDB, err := openDB(dsn)
	if err != nil {
		errLog.Fatal(err)
	}
	defer sqlDB.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		errLog.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()                  // initializze new session manager
	sessionManager.Store = mysqlstore.New(sqlDB) // configure session store to use MySQL db
	sessionManager.Lifetime = 12 * time.Hour     // session automatically expire 12 hours later

	sessionManager.Cookie.Secure = true // cookie only be sent by user when HTTPS connection is being used

	app := &application{
		infoLog:        infoLog,
		errorLog:       errLog,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      app.routes(),
		ErrorLog:     errLog,
		IdleTimeout:  time.Minute, // time to keep-alives connections without repeat hand-shake
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second, // HTTP: time after read finish request header ; HTTPS: time after request is first accepted
	}

	infoLog.Print("Starting server on ", port)

	err = srv.ListenAndServe()
	errLog.Fatal(err)
}

func openDB(dsn string) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err = sqlDB.Ping(); err != nil {
		return nil, nil, err
	}
	return db, sqlDB, nil
}
