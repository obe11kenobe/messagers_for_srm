# messagers_for_crm

Учебный pet-проект: чат-мессенджер на Go для CRM. Пользователи создают
диалоги, обмениваются сообщениями, новые сообщения доставляются в реальном
времени через WebSocket.

## Стек

- Go, стандартный `net/http`
- SQLite (`modernc.org/sqlite`) — БД в одном файле, без отдельного сервера
- WebSocket (`nhooyr.io/websocket`) — доставка сообщений в реальном времени

## Запуск

```bash
go run main.go
```

Сервер поднимется на `localhost:8080`, файл БД `messenger.db` создастся
рядом с проектом при первом запуске.

## API

- `POST /conversations` — создать диалог, тело `{"user_ids": [1, 2]}`
- `GET /conversations/{id}/messages` — история сообщений диалога
- `GET /conversations/{id}/ws` — подключение по WebSocket, realtime-доставка новых сообщений
- `GET /messages` — все сообщения
- `POST /messages` — отправить сообщение, тело `{"conversation_id": 1, "sender_id": 1, "text": "..."}`

## Структура

```
main.go            — сборка приложения, регистрация маршрутов
messager/store.go  — модели и работа с БД
messager/handler.go — HTTP-хендлеры
messager/ws.go      — hub для WebSocket-рассылки
```

Подробный план и пошаговые заметки по разработке — см. `PLAN.md`.
