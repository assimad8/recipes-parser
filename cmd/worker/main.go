package main

import (
	"context"
	"errors"
	"log"
	"time"

	"parser/internal/domain/entities"
	"parser/internal/domain/valueobjects"
	"parser/internal/infrastructure/cache/redis"
	"parser/internal/infrastructure/database/mongo"
	"parser/internal/infrastructure/parser/reddit"
	"parser/internal/infrastructure/queue/rabbitmq"
)

func main() {
	// --------------------
	// Redis
	// --------------------
	redisClient := redis.NewClient(
		"localhost:6379",
		"password",
		0,
		1*time.Hour,
	)
	cache := redis.NewCache(redisClient)

	// JobGuard 
	jobGuard := redis.NewJobGuard(redisClient.Raw(),15*time.Minute)

	// --------------------
	// MongoDB
	// --------------------
	mongoClient, err := mongo.NewClient(
		"mongodb://root:password@localhost:27017",
		"recipesdb",
		"recipes",
	)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}
	db := mongo.NewRepository(mongoClient)

	// --------------------
	// RabbitMQ
	// --------------------
	rmqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		"recipes_jobs",
	)
	if err != nil {
		log.Fatal("RabbitMQ connection error:", err)
	}
	queue := rabbitmq.NewQueue(rmqClient)

	// --------------------
	// Parser (Reddit)
	// --------------------
	parser := reddit.NewRedditParser()

	// --------------------
	// Consume jobs
	// --------------------
	log.Println("Worker started, waiting for jobs...")

	err = queue.Consume(func(job valueobjects.Job) error {
		ctx := context.Background()
		log.Println("Processing job:", job.URL)
	
		defer func() {
			if err := jobGuard.Release(ctx, job.Key, job.Token); err != nil {
				log.Println("Failed to release lock:", err)
			}
		}()
	
		recipes, err := parser.FetchRecipes(job.URL, job.Limit)
		if err != nil {
			switch {
			case errors.Is(err, reddit.ErrNotFound), errors.Is(err, reddit.ErrInvalidURL):
				log.Println("Permanent error, not retrying:", job.URL)
				_ = db.Save(job.Key, []entities.Recipe{}) // save empty result to prevent requeue
				return nil // ACK in RabbitMQ, do not requeue
	
			case errors.Is(err, reddit.ErrRateLimited):
				log.Println("Rate limited, retrying:", job.URL)
				return err // retry
	
			default:
				log.Println("Unknown error, retrying:", err)
				return err // retry
			}
		}
	
		// Success path
		if err := db.Save(job.Key, recipes); err != nil {
			log.Println("Mongo save error:", err)
			return err
		}
	
		if err := cache.Set(job.Key, recipes); err != nil {
			log.Println("Redis cache error:", err)
		}
	
		log.Println("Job completed:", job.URL)
		return nil
	})
	
	
	
	if err != nil {
		log.Fatal("Consume error:", err)
	}

	select {} // keep running
}
