package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
}

type service struct {
	db *mongo.Client
}

var (
	appName    = os.Getenv("APP_NAME")
	dbUserName = os.Getenv("DB_USERNAME")
	dbPassword = os.Getenv("DB_USER_PASSWORD")
	dbHost     = os.Getenv("DB_HOST")
)

func New() Service {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	log.Println(appName, dbUserName, dbPassword, dbHost)
	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=%s",
		dbUserName,
		dbPassword,
		dbHost,
		appName,
	)
	log.Println(uri)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(
		context.Background(),
		opts,
	)

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal("failed to disconnect from mongodb: ", err)
		}
	}()

	if err != nil {
		log.Fatal(err)

	}
	
	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Send a ping to confirm a successful connection
	if err := s.db.Database("admin").RunCommand(
		ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		log.Fatalf("db down: %v", err)
	}
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")

	return map[string]string{
		"message": "It's healthy",
	}
}
