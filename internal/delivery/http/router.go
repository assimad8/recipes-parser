package http

import (
	"parser/internal/delivery/http/handlers"

	"github.com/gin-gonic/gin"
)

func NewRouter(recipeHandler *handlers.RecipeHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		api.POST("/recipes",recipeHandler.PostRecipes)
	}
	return r
}
