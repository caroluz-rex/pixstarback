package main

import (
	"encoding/json"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Pixel struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

type Message struct {
	Type  string `json:"type"`
	Pixel Pixel  `json:"pixel"`
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return "#" + hexByte(uint8(r>>8)) + hexByte(uint8(g>>8)) + hexByte(uint8(b>>8))
}

func hexByte(b uint8) string {
	const hex = "0123456789ABCDEF"
	return string([]byte{hex[b>>4], hex[b&0x0F]})
}

func loadPixelsFromImage(filename string, offsetX, offsetY int) ([]Pixel, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	var pixels []Pixel
	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			colorHex := colorToHex(c)

			if colorHex != "#000000" {
				pixels = append(pixels, Pixel{
					X:     x + offsetX,
					Y:     y + offsetY,
					Color: colorHex,
				})
			}
		}
	}

	return pixels, nil
}

func shufflePixels(pixels []Pixel) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(pixels), func(i, j int) { pixels[i], pixels[j] = pixels[j], pixels[i] })
}

func generateRandomPixels(width, height, count int) []Pixel {
	rand.Seed(time.Now().UnixNano())
	var pixels []Pixel

	for i := 0; i < count; i++ {
		x := rand.Intn(width)
		y := rand.Intn(height)
		colorHex := colorToHex(color.RGBA{
			R: uint8(rand.Intn(256)),
			G: uint8(rand.Intn(256)),
			B: uint8(rand.Intn(256)),
			A: 255,
		})

		pixels = append(pixels, Pixel{
			X:     x,
			Y:     y,
			Color: colorHex,
		})
	}

	return pixels
}

func main() {
	serverURL := "ws://localhost:8080/ws/send"
	imageFiles := []struct {
		Filename string
		OffsetX  int
		OffsetY  int
	}{
		{"bot/pxArt.png", 200, 100},
		{"bot/pxArt3.png", 300, 50},
	}

	var allPixels []Pixel
	for _, imageFile := range imageFiles {
		pixels, err := loadPixelsFromImage(imageFile.Filename, imageFile.OffsetX, imageFile.OffsetY)
		if err != nil {
			log.Printf("Error loading image %s: %v\n", imageFile.Filename, err)
			continue
		}
		allPixels = append(allPixels, pixels...)
	}

	// Добавление случайных пикселей
	randomPixels := generateRandomPixels(500, 300, 1000) // Указываем размеры канваса и количество случайных пикселей
	allPixels = append(allPixels, randomPixels...)

	// Перемешивание всех пикселей
	shufflePixels(allPixels)

	u, err := url.Parse(serverURL)
	if err != nil {
		log.Fatal("Invalid WebSocket URL:", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
	}
	defer conn.Close()

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}
			log.Printf("Received: %s\n", message)
		}
	}()

	for _, pixel := range allPixels {
		message := Message{
			Type:  "update",
			Pixel: pixel,
		}

		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Println("Error marshaling message:", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}

		time.Sleep(5 * time.Millisecond)
	}

	log.Println("Bot finished painting!")
}
