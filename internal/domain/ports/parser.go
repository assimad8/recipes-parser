package ports

import "parser/internal/domain/entities"

type Parser interface {
	Parse(data []byte, limit int) ([]entities.Recipe,error)
}
