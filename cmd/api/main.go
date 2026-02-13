package main
// "https://www.reddit.com/r/recipes/new"
import (
	"log"
	"parser/internal/application/usecases"
	"parser/internal/delivery/http"
	"parser/internal/delivery/http/handlers"
	"parser/internal/infrastructure/cache/redis"
	"parser/internal/infrastructure/database/mongo"
	"parser/internal/infrastructure/parser/reddit"
	"parser/internal/infrastructure/queue/rabbitmq"
	"time"
)

func main() {
	// Redis
	redisClient := redis.NewClient(
		"localhost:6379",
		"password",
		0,
		1*time.Hour,
	)
	cache := redis.NewCache(redisClient)

	// Mongo
	mongoClient,err := mongo.NewClient(
		"mongodb://root:password@localhost:27017/recipesdb?authSource=admin",
		"recipesdb",
		"recipes",
	)
	if err!=nil {
		log.Fatal(err)
	}
	
	db := mongo.NewRepository(mongoClient)
	
	// RabbitMQ
	rmqClient, err := rabbitmq.NewClient(
		"amqp://guest:guest@localhost:5672/",
		"recipes_jobs",
	)
	if err != nil {
		log.Fatal("RabbitMQ connection error:", err)
	}

	queue := rabbitmq.NewQueue(rmqClient)

	redditParser := reddit.NewRedditParser()

	jobGuard := redis.NewJobGuard(redisClient.Raw(), 15*time.Minute)

	usecase := usecases.NewRecipeUseCase(
		cache,
		db,
		queue,
		jobGuard,
		redditParser,
	)

	// HTTP
	handler := handlers.NewRecipeHandler(usecase)
	router := http.NewRouter(handler)

	log.Println("API running on :8080")
	router.Run(":8080")
}