package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"telebridge/internal/api"
	"telebridge/internal/domain"
	"telebridge/internal/queue"
	"telebridge/internal/repository"
	"telebridge/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// لود کردن کانفیگ
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// اتصال به دیتابیس و مهاجرت جداول
	db := database.InitPostgres(os.Getenv("DB_DSN"))
	db.AutoMigrate(&domain.Bot{})

	// اتصال به RabbitMQ
	q, err := queue.NewRabbitMQ(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatal("❌ RabbitMQ connection error:", err)
	}
	defer q.Close()

	// راه‌اندازی لایه‌ها
	botRepo := repository.NewBotRepository(db)
	h := api.NewHandler(botRepo, q)

	// تنظیمات Router
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		// لاراول از این مسیر برای ثبت بات استفاده می‌کند
		v1.POST("/bots", api.AuthMiddleware(), h.RegisterBot)
		// تلگرام پیام‌ها را به این مسیر می‌فرستد
		v1.POST("/webhook/:slug", h.HandleWebhook)
	}

	// تنظیم سرور HTTP
	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	// اجرای سرور در یک Goroutine
	go func() {
		log.Printf("🚀 Server running on port %s", os.Getenv("PORT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⚠️ Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exited properly")
}
