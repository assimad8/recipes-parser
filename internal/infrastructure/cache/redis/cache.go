package redis

import (
	"encoding/json"
	"fmt"
	"parser/internal/domain/entities"
	"parser/internal/domain/ports"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *Client
}

var _ ports.Cache = (*Cache)(nil) //compile-time check


func NewCache(client *Client) *Cache {
	return &Cache{
		client: client,
	}
}


func (c *Cache) Get(key string)([]entities.Recipe,bool,error){
	val,err := c.client.db.Get(c.client.context(),key).Result()
	if err!=nil {
		if err == redis.Nil {
			return nil,false,nil // cache miss
		}
		return nil,false,err
	}
	fmt.Println("val from cache",val)
	var recipes []entities.Recipe
	if err:= json.Unmarshal([]byte(val),&recipes);err!=nil {
		return nil,false,err
	}

	return recipes ,true,nil
}

func (c *Cache) Set(key string, value []entities.Recipe) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal recipes: %w", err)
	}

	err = c.client.db.Set(
		c.client.context(),
		key,
		data,
		c.client.ttl,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to set key in redis: %w", err)
	}

	return nil
}
