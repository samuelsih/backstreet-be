package repo

import (
	"backstreetlinkv2/cmd/model"
	"context"

	_ "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)


func InsertLink[T model.ShortenRequest | model.ShortenFileRequest](ctx context.Context, client *mongo.Client, data T) error {
	db := client.Database("backstreet")
	collection := db.Collection("shorten_link")

	_, err := collection.InsertOne(ctx, data)
	return err
}

func Find(ctx context.Context, client *mongo.Client, param string) (model.ShortenResponse, error) {
	var shorten model.ShortenResponse

	db := client.Database("backstreet")
	collection := db.Collection("shorten_link")

	res := collection.FindOne(ctx, bson.M{"_id": param})
	if res == nil {
		return shorten, mongo.ErrNoDocuments
	}

	if err := res.Decode(&shorten); err != nil {
		return shorten, err
	}

	return shorten, nil
}