package database

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client       // MongoDB client connection (used to talk to the server)
var DB *mongo.Database         // Reference to a specific database within MongoDB

// Init loads env, connects to MongoDB, pings, selects DB, and (optionally) logs collections.
func Init() *mongo.Client {
	// Load .env (non-fatal if missing; system envs still work)
	if rootPath, err := os.Getwd(); err == nil {
		_ = godotenv.Load(filepath.Join(rootPath, ".env"))
	} else {
		log.Println("could not get working directory to load .env:", err)
	}

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI not set in environment")
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		log.Fatal("MONGO_DB_NAME not set in environment")
	}

	// Timeout context for connect + ping
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal("Mongo connect error:", err)
	}
	
	log.Println("Trying to connect to MongoDB at", uri)

	// Verify connectivity
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("Mongo ping failed:", err)
	}

	// Select the database (no new connection; just a handle)
	DB = client.Database(dbName)
	Client = client
	log.Println("Successfully connected to MongoDB; DB =", dbName)

	// Optional: list collections for visibility (can remove in prod)
	if names, err := DB.ListCollectionNames(ctx, struct{}{}); err == nil {
		log.Println("Collections:", names)
	} else {
		log.Println("could not list collections:", err)
	}

	return client
}

// Close cleanly disconnects the client (call on app shutdown).
func Close() {
	if Client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = Client.Disconnect(ctx)
}
