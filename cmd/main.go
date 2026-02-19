package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/banking-superapp/analytics-service/config"
	"github.com/banking-superapp/analytics-service/handler"
	"github.com/banking-superapp/analytics-service/repository"
	"github.com/banking-superapp/analytics-service/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

func main() {
	cfg := config.Load()
	mongoClient, err := repository.NewMongoClient(cfg.MongoAtlasURI)
	if err != nil {
		log.Fatalf("MongoDB connection failed: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	db := mongoClient.Database("banking_analytics")
	if err := repository.CreateIndexes(db); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	eventRepo := repository.NewEventRepo(db)
	segmentRepo := repository.NewSegmentRepo(db)
	crossSellRepo := repository.NewCrossSellRuleRepo(db)
	analyticsSvc := service.NewAnalyticsService(eventRepo, segmentRepo, crossSellRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsSvc)

	app := fiber.New(fiber.Config{AppName: cfg.ServiceName, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second})
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": cfg.ServiceName})
	})

	v1 := app.Group("/v1")
	analytics := v1.Group("/analytics")
	analytics.Post("/event", analyticsHandler.RecordEvent)
	analytics.Get("/segment", analyticsHandler.GetSegment)
	analytics.Get("/crosssell", analyticsHandler.GetCrossSellOffers)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("Starting %s on port %s", cfg.ServiceName, cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
}
