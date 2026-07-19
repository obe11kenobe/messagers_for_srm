package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Conversation struct {
	ID        int
	CreatedAt time.Time
}

type Participant struct {
	ConversationID int
	UserID         int
}

type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderID       int       `json:"sender_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

func getMessages(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, conversation_id, sender_id, text, created_at FROM messages")
		if err != nil {
			http.Error(w, "Ошибка запроса к БД: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var result []Message

		for rows.Next() {
			var m Message
			var createdAt string
			err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Text, &createdAt)
			if err != nil {
				http.Error(w, "Ошибка сканирования строки: "+err.Error(), http.StatusInternalServerError)
				return
			}

			m.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
			if err != nil {
				http.Error(w, "Ошибка разбора времени: "+err.Error(), http.StatusInternalServerError)
				return
			}

			result = append(result, m)
		}

		if err = rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func postMessage(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg Message

		err := json.NewDecoder(r.Body).Decode(&msg)
		if err != nil {
			http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		msg.CreatedAt = time.Now()

		res, err := db.Exec(
			"INSERT INTO messages (conversation_id, sender_id, text, created_at) VALUES (?, ?, ?, ?)",
			msg.ConversationID, msg.SenderID, msg.Text, msg.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Ошибка сохранения в БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := res.LastInsertId()
		if err != nil {
			http.Error(w, "Ошибка получения ID: "+err.Error(), http.StatusInternalServerError)
			return
		}
		msg.ID = int(id)

		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(msg)
		if err != nil {
			http.Error(w, "Ошибка при кодировании ответа: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func main() {

	db, err := sql.Open("sqlite", "messenger.db")
	if err != nil {
		log.Fatal("Ошибка открытия БД:", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		created_at TEXT
	)`)

	if err != nil {
		log.Fatal("Ошибка создания таблицы conversations:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS participants (
		conversation_id INTEGER, 
		user_id INTEGER
	)`)

	if err != nil {
		log.Fatal("Ошибка при создании таблицы participants:", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER,
    sender_id INTEGER,
    text TEXT,
    created_at TEXT
	)`)

	if err != nil {
		log.Fatal("Ошибки при создании таблицы messages:", err)
	}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getMessages(db)(w, r)
		case "POST":
			postMessage(db)(w, r)
		}
	})

	log.Print("Сервер запущен на порту 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
