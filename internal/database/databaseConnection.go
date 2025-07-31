package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client = dbInstance()

func dbInstance() *mongo.Client {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("error loading the env file...")
	}

	MongoDB := os.Getenv("MONGO_URL")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal("Mongo DB Client Error", err)
	}

	log.Println("Succesfully connected to MongoDB!")

	return client
}
