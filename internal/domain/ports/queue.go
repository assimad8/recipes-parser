package ports

import "parser/internal/domain/valueobjects"

type Queue interface {
	Publish(job valueobjects.Job) error
	Consume(handler func(valueobjects.Job) error) error
}
