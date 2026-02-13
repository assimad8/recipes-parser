package mongo

import (
	"context"
	"time"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	db *mongodriver.Database
	collection *mongodriver.Collection
}

func NewClient(uri,dbName,collection string) (*Client,error){
	ctx,cancel := context.WithTimeout(context.Background(),10*time.Second)
	defer cancel()

	client,err := mongodriver.Connect(ctx,options.Client().ApplyURI(uri))
	if err!=nil {
		return nil,err
	}

	return &Client{
		db: client.Database(dbName),
		collection: client.Database(dbName).Collection(collection),
	},nil
}

func (c *Client)context()(context.Context, context.CancelFunc){
	return context.WithTimeout(context.Background(),5*time.Second)
}