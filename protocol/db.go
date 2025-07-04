package protocol

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	db *mongo.Client
)

func RegisterMongoDB(client *mongo.Client) {
	db = client
}

func RegisterMongoDBByURI(uri string) error {
	var err error
	db, err = mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	return err
}
