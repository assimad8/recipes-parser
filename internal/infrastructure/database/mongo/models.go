package mongo

import (
	"parser/internal/domain/entities"
	"time"
)

type RecipeDocument struct {
	ID        string            `bson:"_id"`
	Recipes   []entities.Recipe `bson:"recipes"`
	CreatedAt time.Time         `bson:"created_at"`
}
