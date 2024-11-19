package websocket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
	"your_project/repositories"

	"go.mongodb.org/mongo-driver/mongo"
	"your_project/services"
)

type Hub struct {
	sendClients       map[*Client]bool // Клиенты для отправки
	receiveClients    map[*Client]bool // Клиенты для получения
	broadcast         chan []byte
	registerSend      chan *Client
	registerReceive   chan *Client
	unregisterSend    chan *Client
	unregisterReceive chan *Client
	pixelService      services.PixelService
	Logger            *log.Logger
	mutex             sync.RWMutex
}

func NewHub(mongoClient *mongo.Client, dbName string) *Hub {
	db := mongoClient.Database(dbName)
	pixelRepo := repositories.NewPixelRepository(db)
	pixelService := services.NewPixelService(pixelRepo)

	return &Hub{
		sendClients:       make(map[*Client]bool),
		receiveClients:    make(map[*Client]bool),
		broadcast:         make(chan []byte),
		registerSend:      make(chan *Client),
		registerReceive:   make(chan *Client),
		unregisterSend:    make(chan *Client),
		unregisterReceive: make(chan *Client),
		pixelService:      pixelService,
		Logger:            log.Default(),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.registerSend:
			h.mutex.Lock()
			h.sendClients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregisterSend:
			h.mutex.Lock()
			if _, ok := h.sendClients[client]; ok {
				delete(h.sendClients, client)
				close(client.send)
			}
			h.mutex.Unlock()
		case client := <-h.registerReceive:
			h.mutex.Lock()
			h.receiveClients[client] = true
			h.mutex.Unlock()
			go h.sendInitialState(client)
		case client := <-h.unregisterReceive:
			h.mutex.Lock()
			if _, ok := h.receiveClients[client]; ok {
				delete(h.receiveClients, client)
				close(client.send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.receiveClients { // Только клиенты для получения
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.receiveClients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) RegisterSendClient(client *Client) {
	h.registerSend <- client
}

func (h *Hub) RegisterReceiveClient(client *Client) {
	h.registerReceive <- client
}

func (h *Hub) UnregisterSendClient(client *Client) {
	h.unregisterSend <- client
}

func (h *Hub) UnregisterReceiveClient(client *Client) {
	h.unregisterReceive <- client
}

func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

func (h *Hub) sendInitialState(client *Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pixels, err := h.pixelService.GetAllPixels(ctx)
	if err != nil {
		h.Logger.Println("Error fetching initial pixels:", err)
		return
	}

	initialMessage := map[string]interface{}{
		"type":   "initial",
		"pixels": pixels,
	}

	message, err := json.Marshal(initialMessage)
	if err != nil {
		h.Logger.Println("Error marshaling initial message:", err)
		return
	}

	client.send <- message
}
