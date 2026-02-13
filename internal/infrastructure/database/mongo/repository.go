package mongo

import (
	"errors"
	"parser/internal/domain/entities"
	"parser/internal/domain/ports"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
	client *Client
}

var _ ports.DataBase = (*Repository)(nil)

func NewRepository(client *Client) *Repository {
	return &Repository{
		client: client,
	}
}

func (r *Repository) Find(key string)([]entities.Recipe,bool,error){
	ctx ,cancel := r.client.context()
	defer cancel()

	var doc RecipeDocument

	err := r.client.collection.
			FindOne(ctx,bson.M{"_id":key}).
			Decode(&doc)
	if err!=nil {
		if errors.Is(err,mongo.ErrNoDocuments) {
			return nil,false,nil
		}
		return nil,false,err
	}

	return doc.Recipes,true,nil
}

func (r *Repository) Save(key string, recipes []entities.Recipe) error {
	ctx, cancel := r.client.context()
	defer cancel()

	now := time.Now()

	update := bson.M{
		"$set": bson.M{
			"recipes":   recipes,
			"updatedAt": now,
		},
		"$setOnInsert": bson.M{
			"_id":       key,
			"createdAt": now,
		},
	}

	_, err := r.client.collection.UpdateOne(
		ctx,
		bson.M{"_id": key},
		update,
		mongoDriverOptions(),
	)

	return err
}


func mongoDriverOptions() *options.UpdateOptions {
	upsert := true
	return &options.UpdateOptions{
		Upsert: &upsert,
	}
}