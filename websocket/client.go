package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"your_project/models"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	conn *websocket.Conn
	hub  *Hub
	send chan []byte
}

func NewSendClient(conn *websocket.Conn, hub *Hub) *Client {
	client := &Client{
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 256),
	}

	go client.readPump() // Чтение данных
	return client
}

func NewReceiveClient(conn *websocket.Conn, hub *Hub) *Client {
	client := &Client{
		conn: conn,
		hub:  hub,
		send: make(chan []byte, 256),
	}

	go client.writePump() // Отправка данных клиенту
	return client
}

func (c *Client) readPump() {
	defer func() {
		// Убедитесь, что используется правильный метод для снятия регистрации
		if c.hub != nil {
			if _, isSendClient := c.hub.sendClients[c]; isSendClient {
				c.hub.UnregisterSendClient(c)
			} else if _, isReceiveClient := c.hub.receiveClients[c]; isReceiveClient {
				c.hub.UnregisterReceiveClient(c)
			}
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.Logger.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Обработка входящих сообщений
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			c.hub.Logger.Println("Error unmarshaling message:", err)
			continue
		}

		if msg["type"] == "update" {
			pixelData := msg["pixel"].(map[string]interface{})
			pixel := models.Pixel{
				X:     int(pixelData["x"].(float64)),
				Y:     int(pixelData["y"].(float64)),
				Color: pixelData["color"].(string),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := c.hub.pixelService.UpsertPixel(ctx, pixel); err != nil {
				c.hub.Logger.Println("Error upserting pixel:", err)
				continue
			}

			updateMessage := map[string]interface{}{
				"type":  "update",
				"pixel": pixel,
			}

			message, err := json.Marshal(updateMessage)
			if err != nil {
				c.hub.Logger.Println("Error marshaling update message:", err)
				continue
			}

			c.hub.Broadcast(message)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		// Убедитесь, что используется правильный метод для снятия регистрации
		if c.hub != nil {
			if _, isSendClient := c.hub.sendClients[c]; isSendClient {
				c.hub.UnregisterSendClient(c)
			} else if _, isReceiveClient := c.hub.receiveClients[c]; isReceiveClient {
				c.hub.UnregisterReceiveClient(c)
			}
		}
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Канал закрыт
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			writer.Write(message)

			// Добавляем очередные сообщения
			n := len(c.send)
			for i := 0; i < n; i++ {
				writer.Write([]byte{'\n'})
				writer.Write(<-c.send)
			}

			if err := writer.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
