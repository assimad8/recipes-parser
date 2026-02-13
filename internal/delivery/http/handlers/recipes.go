package handlers

import (
	"net/http"
	"parser/internal/application/usecases"
	"parser/internal/domain/entities"
	"time"

	"github.com/gin-gonic/gin"
)

type createRecipeRequest struct {
	URL   string `json:"url" binding:"required,url"`
	Limit int    `json:"limit"`
}

type recipeResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	CreatedAt string `json:"created_at"`
	ImageURL string `json:"image_url"`
	Link     string `json:"link"`
}

type RecipeHandler struct {
	usecase usecases.RecipeUseCase
}

func NewRecipeHandler(uc usecases.RecipeUseCase) *RecipeHandler {
	return &RecipeHandler{
		usecase: uc,
	}
}

func (h *RecipeHandler) PostRecipes(c *gin.Context) {
	var req createRecipeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	result, err := h.usecase.GetRecipes(
		c.Request.Context(),
		req.URL,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	switch result.Status {

	case usecases.StatusFound:
		data,count := toHTTPRecipes(result.Recipes)
		c.JSON(http.StatusOK, gin.H{
			"status": "found",
			"data": data,
			"count":count,
		})
		return

	case usecases.StatusProcessing:
		c.Header("Retry-After", "5")
		c.JSON(http.StatusAccepted, gin.H{
			"status": "processing",
		})
		return
	}
}

func toHTTPRecipes(recipes []entities.Recipe) ([]recipeResponse,int) {
	out := make([]recipeResponse, 0, len(recipes))
	count :=0
	for _, r := range recipes {
		out = append(out, recipeResponse{
			ID:        r.ID,
			Title:     r.Title,
			Author:    r.Author,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
			ImageURL:  r.ImageURL,
			Link:      r.Link,
		})
		count++
	}

	return out,count
}
