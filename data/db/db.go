package db

import (
	"context"
	"github.com/han-so1omon/concierge.map/util"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var Client *mongo.Client

func ConnectDatabase() {
	var err error
	mongodbCxnString := os.Getenv("MONGODB_CONNECTION_STRING")
	util.Logger.Info("Database connecting...")
	clientOptions := options.Client().ApplyURI(mongodbCxnString)

	Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		util.Logger.Fatal(err)
	}

	err = Client.Ping(context.TODO(), nil)
	if err != nil {
		util.Logger.Fatal(err)
	}

	util.Logger.Info("Database connected")
}
