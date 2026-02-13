package ports

import "parser/internal/domain/entities"

type Cache interface {
	Get(key string) ([]entities.Recipe,bool,error)
	Set(key string,value []entities.Recipe) error
}
