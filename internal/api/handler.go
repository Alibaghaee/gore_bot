package api

import (
	"net/http"
	"telebridge/internal/domain"
	"telebridge/internal/queue"
	"telebridge/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Repo  *repository.BotRepository
	Queue *queue.RabbitMQ
}

// این همان تابعی است که main.go به دنبال آن می‌گردد
func NewHandler(repo *repository.BotRepository, q *queue.RabbitMQ) *Handler {
	return &Handler{
		Repo:  repo,
		Queue: q,
	}
}

func (h *Handler) RegisterBot(c *gin.Context) {
	var bot domain.Bot
	if err := c.ShouldBindJSON(&bot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	if err := h.Repo.Create(c.Request.Context(), &bot); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bot)
}

func (h *Handler) HandleWebhook(c *gin.Context) {
	slug := c.Param("slug")
	bot, err := h.Repo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bot not found"})
		return
	}

	var update map[string]interface{}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	go h.Queue.Publish(map[string]interface{}{
		"bot_id": bot.ID,
		"data":   update,
	})

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}
