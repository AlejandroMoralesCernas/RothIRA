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

var Client *mongo.Client
var DB *mongo.Database

func Init() {
	// Load .env (ignore error if already loaded)
	_ = godotenv.Load(".env")

	uri := os.Getenv("MONGODB_URI") // matches your .env
	dbName := os.Getenv("MONGODB_DB")
	if uri == "" || dbName == "" {
		log.Fatal("MONGODB_URI or MONGODB_DB not set in .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}

	Client = client
	DB = client.Database(dbName)
	log.Println("Successfully connected to MongoDB:", dbName)
}
