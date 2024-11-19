package controllers

import (
	"net/http"

	"your_project/websocket"
)

// HandleSendWebSocket - WebSocket для отправки пикселей
func HandleSendWebSocket(hub *websocket.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.Logger.Println("WebSocket send upgrade error:", err)
		return
	}

	client := websocket.NewSendClient(conn, hub) // Новый клиент для отправки
	hub.RegisterSendClient(client)
}

// HandleReceiveWebSocket - WebSocket для получения обновлений
func HandleReceiveWebSocket(hub *websocket.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.Logger.Println("WebSocket receive upgrade error:", err)
		return
	}

	client := websocket.NewReceiveClient(conn, hub) // Новый клиент для получения
	hub.RegisterReceiveClient(client)
}
