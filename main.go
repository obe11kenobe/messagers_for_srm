package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"messagers_for_crm/messager"

	_ "modernc.org/sqlite"
)

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("Не задана переменная окружения JWT_SECRET")
	}

	db, err := sql.Open("sqlite", "messenger.db")
	if err != nil {
		log.Fatal("Ошибка открытия БД:", err)
	}
	defer db.Close()

	store := messager.NewStore(db)
	if err := store.CreateTables(); err != nil {
		log.Fatal("Ошибка создания таблиц:", err)
	}

	h := messager.NewHub()

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			messager.RequireAuth(secret, messager.GetMessages(store))(w, r)
		case "POST":
			messager.RequireAuth(secret, messager.PostMessage(store, h))(w, r)
		}
	})

	http.HandleFunc("/conversations", messager.RequireAuth(secret, messager.CreateConversation(store)))

	http.HandleFunc("/conversations/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ws") {
			messager.WSHandler(h, secret)(w, r)
			return
		}
		messager.RequireAuth(secret, messager.ConversationMessages(store))(w, r)
	})

	log.Print("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", logRequests(http.DefaultServeMux)))
}
