package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	models "workout-manager/internal/models/database/user"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IDatabase interface {
	Disconnect(ctx context.Context) error
	Collection(name string) *mongo.Collection
	Health(ctx context.Context) map[string]string
}

type Database struct {
	client *mongo.Client
	IDatabase
}

var (
	appName    = os.Getenv("APP_NAME")
	dbUserName = os.Getenv("DB_USERNAME")
	dbPassword = os.Getenv("DB_USER_PASSWORD")
	dbHost     = os.Getenv("DB_HOST")
)

func New() *Database {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	uri := fmt.Sprintf(
		"mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=%s",
		dbUserName,
		dbPassword,
		dbHost,
		appName,
	)

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(
		context.Background(),
		opts,
	)
	if err != nil {
		log.Fatal(err)

	}

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.Background(), bson.D{{"ping", 1}}).Err(); err != nil {
		log.Fatal("failed to connect to database: ", err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	return &Database{
		client: client,
	}
}

func (d *Database) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Send a ping to confirm a successful connection
	if err := d.client.Database("admin").RunCommand(
		ctx, bson.D{{"ping", 1}}).Err(); err != nil {
		log.Fatalf("db down: %v", err)
	}
	return map[string]string{
		"message": "It's healthy",
	}
}

func (d *Database) Disconnect(ctx context.Context) error {
	if err := d.client.Disconnect(ctx); err != nil {
		return err
	}
	return nil
}

func (d *Database) Collection(name string) *mongo.Collection {
	collection := d.client.Database(appName).Collection(name)
	return collection
}

func (d *Database) Register(user models.User) error {
	collection := d.client.Database(appName).Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем, существует ли пользователь
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	//if err != nil {
	//	log.Printf("failed to find existing user: %v", err)
	//	return err
	//}
	if existingUser.Username == user.Username {
		return ErrorUserExist
	}

	// Добавляем пользователя в БД
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("failed to insert user: %v", err)
		return ErrorUserInsert
	}
	return nil
}

func (d *Database) Login(username, password string) (*models.User, error) {
	collection := d.client.Database(appName).Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		return nil, ErrorUserNotFound
	case err != nil:
		return &user, ErrorSomethingGetWrong
	}

	return &user, nil

}
