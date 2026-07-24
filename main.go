package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"messagers_for_crm/messager"

	_ "modernc.org/sqlite"
)

func main() {
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
			messager.GetMessages(store)(w, r)
		case "POST":
			messager.PostMessage(store, h)(w, r)
		}
	})

	http.HandleFunc("/conversations/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ws") {
			messager.WSHandler(h)(w, r)
			return
		}
		messager.ConversationMessages(store)(w, r)
	})

	http.HandleFunc("/conversations", messager.CreateConversation(store))

	log.Print("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
