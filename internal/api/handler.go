package api

import (
	"fmt"
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

	// Go اینجا یک بار کوئری زده و همه دیتای بات را دارد
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

	// ساخت Routing Key هوشمند: telegram.message.{owner_id}.{plan_type}
	// مثال: telegram.message.42.premium
	routingKey := fmt.Sprintf("telegram.message.%d.%s", bot.OwnerID, bot.PlanType)

	// تزریق تمام دیتای لازم برای لاراول
	payload := map[string]interface{}{
		"bot_id":   bot.ID,
		"owner_id": bot.OwnerID,
		"token":    bot.Token,    // لاراول برای پاسخ دادن به تلگرام به این نیاز دارد
		"plan":     bot.PlanType, // برای اولویت‌بندی در سمت لاراول
		"payload":  update,       // خودِ پیام تلگرام
	}

	// ارسال به Exchange با Routing Key مخصوص
	go func() {
		err := h.Queue.Publish("telegram_events", routingKey, payload)
		if err != nil {
			// در سیستم‌های واقعی اینجا لاگ بزنید
		}
	}()

	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}
