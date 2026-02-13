package usecases

import "context"

type RecipeUseCase interface {
	GetRecipes(context.Context,string,int) (Result, error)
}