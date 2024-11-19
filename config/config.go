package config

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	ServerAddress string
	MongoURI      string
	DatabaseName  string
	MongoUser     string
	MongoPassword string
}

func LoadConfig() *Config {
	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),
		MongoURI:      getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:  getEnv("DATABASE_NAME", "pixelcanvas"),
		MongoUser:     getEnv("MONGO_USER", "admin"),
		MongoPassword: getEnv("MONGO_PASSWORD", "password"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func InitMongoDB(uri, username, password string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri).SetAuth(options.Credential{
		Username: username,
		Password: password,
	})

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}

	log.Println("Connected to MongoDB")
	return client, nil
}
