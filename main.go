// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"your_project/repositories"
	"your_project/services"

	"your_project/config"
	"your_project/controllers"
	"your_project/middlewares"
	"your_project/websocket"
)

func main() {
	// Загрузка конфигурации
	cfg := config.LoadConfig()

	// Инициализация клиента MongoDB
	mongoClient, err := config.InitMongoDB(cfg.MongoURI, "admin", "admin")
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.Background()); err != nil {
			log.Println("Error disconnecting MongoDB:", err)
		}
	}()

	db := mongoClient.Database(cfg.DatabaseName)
	teamRepo := repositories.NewTeamRepository(db)
	teamService := services.NewTeamService(teamRepo)
	teamController := controllers.NewTeamController(teamService)

	// Инициализация WebSocket хаба
	hub := websocket.NewHub(mongoClient, cfg.DatabaseName)
	go hub.Run()

	// Установка маршрута для WebSocket
	http.Handle("/ws/send", middlewares.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.HandleSendWebSocket(hub, w, r)
	})))

	http.Handle("/ws/receive", middlewares.CORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		controllers.HandleReceiveWebSocket(hub, w, r)
	})))

	// Добавление эндпоинтов аутентификации с использованием CORS middleware
	http.Handle("/get-challenge", middlewares.CORS(http.HandlerFunc(controllers.GetChallengeHandler)))
	http.Handle("/authenticate", middlewares.CORS(http.HandlerFunc(controllers.AuthenticateHandler)))
	http.Handle("/teams", middlewares.CORS(http.HandlerFunc(teamController.GetTeamsHandler)))
	http.Handle("/teams/members", middlewares.CORS(http.HandlerFunc(teamController.GetTeamMembersHandler)))
	http.Handle("/teams/leave", middlewares.CORS(middlewares.JWTAuth(http.HandlerFunc(teamController.LeaveTeamHandler))))
	http.Handle("/me", middlewares.CORS(middlewares.JWTAuth(http.HandlerFunc(controllers.MeHandler))))
	http.Handle("/teams/create", middlewares.CORS(middlewares.JWTAuth(http.HandlerFunc(teamController.CreateTeamHandler))))
	http.Handle("/teams/join", middlewares.CORS(middlewares.JWTAuth(http.HandlerFunc(teamController.JoinTeamHandler))))
	http.Handle("/logout", middlewares.CORS(http.HandlerFunc(controllers.LogoutHandler)))

	// Запуск HTTP-сервера
	log.Printf("Server is running on %s", cfg.ServerAddress)
	if err := http.ListenAndServe(cfg.ServerAddress, nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
