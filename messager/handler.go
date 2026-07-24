package messager

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"nhooyr.io/websocket"
)

func GetMessages(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := s.AllMessages()
		if err != nil {
			http.Error(w, "Ошибка запроса к БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func ConversationMessages(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/conversations/")
		idStr = strings.TrimSuffix(idStr, "/messages")

		convID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Некорректный id диалога", http.StatusBadRequest)
			return
		}

		result, err := s.MessagesByConversation(convID)
		if err != nil {
			http.Error(w, "Ошибка запроса к БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func PostMessage(s *Store, h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var msg Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		exists, err := s.ConversationExists(msg.ConversationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Диалог не найден", http.StatusNotFound)
			return
		}

		msg, err = s.InsertMessage(msg)
		if err != nil {
			http.Error(w, "Ошибка сохранения в БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		h.broadcast(msg.ConversationID, msg)

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(msg); err != nil {
			http.Error(w, "Ошибка при кодировании ответа: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func CreateConversation(s *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Некорректный JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		conv, err := s.CreateConversation(req.UserIDs)
		if err != nil {
			http.Error(w, "Ошибка сохранения в БД: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(conv); err != nil {
			http.Error(w, "Ошибка при кодировании ответа: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

func WSHandler(h *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/conversations/")
		idStr = strings.TrimSuffix(idStr, "/ws")

		convID, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Некорректный id диалога", http.StatusBadRequest)
			return
		}

		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		h.subscribe(convID, c)
		defer h.unsubscribe(convID, c)

		for {
			if _, _, err := c.Read(context.Background()); err != nil {
				return
			}
		}
	}
}
