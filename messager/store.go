package messager

import (
	"database/sql"
	"time"
)

type Conversation struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type Participant struct {
	ConversationID int `json:"conversation_id"`
	UserID         int `json:"user_id"`
}

type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderID       int       `json:"sender_id"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateConversationRequest struct {
	UserIDs []int `json:"user_ids"`
}

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateTables() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at TEXT
	)`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS participants (
		conversation_id INTEGER,
		user_id INTEGER
	)`)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER,
		sender_id INTEGER,
		text TEXT,
		created_at TEXT
	)`)
	return err
}

func scanMessages(rows *sql.Rows) ([]Message, error) {
	var result []Message

	for rows.Next() {
		var m Message
		var createdAt string
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Text, &createdAt); err != nil {
			return nil, err
		}

		parsed, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		m.CreatedAt = parsed

		result = append(result, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Store) AllMessages() ([]Message, error) {
	rows, err := s.db.Query("SELECT id, conversation_id, sender_id, text, created_at FROM messages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (s *Store) MessagesByConversation(convID int) ([]Message, error) {
	rows, err := s.db.Query(
		"SELECT id, conversation_id, sender_id, text, created_at FROM messages WHERE conversation_id = ?",
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessages(rows)
}

func (s *Store) ConversationExists(convID int) (bool, error) {
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM conversations WHERE id = ?)", convID).Scan(&exists)
	return exists, err
}

func (s *Store) InsertMessage(msg Message) (Message, error) {
	msg.CreatedAt = time.Now()

	res, err := s.db.Exec(
		"INSERT INTO messages (conversation_id, sender_id, text, created_at) VALUES (?, ?, ?, ?)",
		msg.ConversationID, msg.SenderID, msg.Text, msg.CreatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return Message{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Message{}, err
	}
	msg.ID = int(id)

	return msg, nil
}

func (s *Store) CreateConversation(userIDs []int) (Conversation, error) {
	createdAt := time.Now()

	res, err := s.db.Exec("INSERT INTO conversations (created_at) VALUES (?)", createdAt.Format(time.RFC3339Nano))
	if err != nil {
		return Conversation{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Conversation{}, err
	}

	for _, userID := range userIDs {
		_, err := s.db.Exec("INSERT INTO participants (conversation_id, user_id) VALUES (?, ?)", id, userID)
		if err != nil {
			return Conversation{}, err
		}
	}

	return Conversation{ID: int(id), CreatedAt: createdAt}, nil
}
