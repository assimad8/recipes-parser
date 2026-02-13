package usecases

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"parser/internal/domain/entities"
	"parser/internal/domain/ports"
	"parser/internal/domain/valueobjects"
	"parser/internal/infrastructure/parser"
)

type Status string

const (
	StatusFound      Status = "FOUND"
	StatusProcessing Status = "PROCESSING"
)

type Result struct {
	Status  Status
	Recipes []entities.Recipe
}

type recipeUseCase struct {
	cache    ports.Cache
	database ports.DataBase
	queue    ports.Queue
	jobGuard ports.JobGuard
	parser   parser.Parser
}

func NewRecipeUseCase(
	cache ports.Cache,
	database ports.DataBase,
	queue ports.Queue,
	jobGuard ports.JobGuard,
	parser parser.Parser,
) RecipeUseCase {
	return &recipeUseCase{
		cache:    cache,
		database: database,
		queue:    queue,
		jobGuard: jobGuard,
		parser:   parser,
	}
}

func (uc *recipeUseCase) GetRecipes(
	ctx context.Context,
	url string,
	limit int,
) (Result, error) {

	if err := validate(url, limit); err != nil {
		return Result{}, err
	}

	key := cacheKey(url, limit)

	// Cache first
	if recipes, found, err := uc.cache.Get(key); err != nil {
		return Result{}, err
	} else if found {
		return Result{
			Status:  StatusFound,
			Recipes: recipes,
		}, nil
	}

	// Database fallback
	if recipes, found, err := uc.database.Find(key); err != nil {
		return Result{}, err
	} else if found {

		if err := uc.cache.Set(key, recipes); err != nil {
			log.Printf("cache set failed: %v", err)
		}

		return Result{
			Status:  StatusFound,
			Recipes: recipes,
		}, nil
	}

	// Acquire distributed lock
	token, acquired, err := uc.jobGuard.Acquire(ctx, key)
	if err != nil {
		return Result{}, err
	}

	if !acquired {
		return Result{Status: StatusProcessing}, nil
	}

	job := valueobjects.Job{
		URL:   url,
		Limit: limit,
		Key:   key,
		Token: token,
	}

	if err := uc.queue.Publish(job); err != nil {
		_ = uc.jobGuard.Release(ctx, key, token)
		return Result{}, err
	}

	return Result{Status: StatusProcessing}, nil
}


func cacheKey(url string, limit int) string {
	h := sha1.Sum([]byte(url))
	return fmt.Sprintf("recipes:%x:%d", h, limit)
}

func validate(url string, limit int) error {
	if url == "" {
		return errors.New("url is required")
	}
	if limit <= 0 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	return nil
}
