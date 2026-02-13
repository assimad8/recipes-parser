package parser

import (
	"parser/internal/domain/entities"
)

type Parser interface {
	FetchRecipes(url string, limit int) ([]entities.Recipe, error)
}
