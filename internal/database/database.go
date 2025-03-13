package database

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"time"
	"workout-manager/internal/models/database/trainings"
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
	appName               = os.Getenv("APP_NAME")
	dbName                = os.Getenv("DB_DATABASE_NAME")
	dbAuthCollection      = os.Getenv("DB_AUTH_COLLECTION")
	dbTrainingsCollection = os.Getenv("DB_TRAININGS_COLLECTION")
	dbUserName            = os.Getenv("DB_USERNAME")
	dbPassword            = os.Getenv("DB_USER_PASSWORD")
	dbHost                = os.Getenv("DB_HOST")
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

//func (d *Database) Collection(name string) *mongo.Collection {
//	collection := d.client.Database(dbName).Collection(name)
//	return collection
//}

func (d *Database) Register(user models.User) error {
	const op = "database/Register"

	collection := d.client.Database(appName).Collection(dbAuthCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем, существует ли пользователь
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
	if existingUser.Username == user.Username {
		log.Printf("%s: User already registered: %s", op, user.Username)
		return ErrorUserExist
	}

	// Добавляем пользователя в БД
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("%s: failed to insert user: %v", op, err)
		return ErrorUserInsert
	}
	return nil
}

func (d *Database) Login(username, password string) (*models.User, error) {
	const op = "database/Login"

	collection := d.client.Database(appName).Collection(dbAuthCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	switch {
	case errors.Is(err, mongo.ErrNoDocuments):
		log.Printf("%s: user %s does not exist", op, user.Username)
		return nil, ErrorUserNotFound
	case err != nil:
		log.Printf("%s: %v", op, err)
		return nil, ErrorSomethingGetWrong
	}

	if user.Password != password {
		log.Printf("%s: Пароли не совпадают", op)
		return nil, ErrorSomethingGetWrong
	}

	return &user, nil

}

func (d *Database) AddTraining(training trainings.Training) (primitive.ObjectID, error) {
	const op = "database/AddTraining"
	collection := d.client.Database(dbName).Collection(dbTrainingsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	training.ID = primitive.NewObjectID()

	_, err := collection.InsertOne(ctx, training)
	if err != nil {
		log.Printf("%s: failed to insert training: %v", op, err)
		return primitive.NilObjectID, err
	}
	return training.ID, nil
}

func (d *Database) DeleteTraining(id primitive.ObjectID) error {
	const op = "database/DeleteTraining"
	collection := d.client.Database(dbName).Collection(dbTrainingsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	num, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Printf("%s: failed to delete training: %v", op, err)
		return err
	}
	if num.DeletedCount == 0 {
		log.Printf("%s: training %s has not been exist", op, id)
		return ErrorTrainingNotExist
	}
	return nil
}

func (d *Database) GetUserTrainings(username string) ([]trainings.Training, error) {
	const op = "database/GetUserTrainings"
	collection := d.client.Database(dbName).Collection(dbTrainingsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var resTrainings []trainings.Training
	cur, err := collection.Find(ctx, bson.M{"username": username})
	if err != nil {
		log.Printf("%s: failed to find trainings: %v", op, err)
		return nil, err
	}
	defer cur.Close(ctx)
	for cur.Next(ctx) {
		var tr trainings.Training
		err := cur.Decode(&tr)
		if err != nil {
			log.Printf("%s: failed to decode training: %v", op, err)
			return nil, err
		}
		resTrainings = append(resTrainings, tr)
	}

	if err := cur.Err(); err != nil {
		log.Printf("%s: failed to decode training: %v", op, err)
		return nil, err
	}

	return resTrainings, nil
}
