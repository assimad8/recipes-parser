package ports

import "parser/internal/domain/entities"

type DataBase interface {
	Find(key string) ([]entities.Recipe,bool, error)
	Save(key string, value []entities.Recipe) error
}
