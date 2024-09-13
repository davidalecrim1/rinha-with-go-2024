package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strconv"
	"time"

	"rinha-with-go-2024/cmd/api/handler"
	"rinha-with-go-2024/cmd/api/handler/middleware"
	"rinha-with-go-2024/config/env"
	"rinha-with-go-2024/internal/domain"
	"rinha-with-go-2024/internal/infra/logger"
	"rinha-with-go-2024/internal/infra/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := initializeLogger()
	db := initializeDatabase()
	monitorConnectionPool(logger, db)

	repo := repository.NewClientRepository(logger, db)
	svc := domain.NewClientRepository(logger, repo)
	handler := handler.NewClientHandler(logger, svc)

	initializeRouter(handler)
}

func initializeLogger() *slog.Logger {
	level := env.GetEnvOrSetDefault("LOG_LEVEL", "DEBUG")
	return logger.NewLogger(level)
}

func initializeDatabase() *pgxpool.Pool {
	dbEndpoint := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.GetEnvOrSetDefault("DB_USER", "admin"),
		env.GetEnvOrSetDefault("DB_PASSWORD", "password"),
		env.GetEnvOrSetDefault("DB_HOST", "localhost"),
		env.GetEnvOrSetDefault("DB_PORT", "5432"),
		env.GetEnvOrSetDefault("DB_SCHEMA", "rinha"))

	config, err := pgxpool.ParseConfig(dbEndpoint)
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	maxConn, err := strconv.Atoi(env.GetEnvOrSetDefault("DB_MAX_CONN", "50"))
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	config.MaxConns = int32(maxConn)
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("error loading database configuration: %v", err)
	}

	return pool
}

func monitorConnectionPool(logger *slog.Logger, db *pgxpool.Pool) {
	enabled := env.GetEnvOrSetDefault("MONITOR_CONN_POOL", "1")

	if enabled != "1" {
		return
	}

	monitor := func(d time.Duration) {
		for {
			stats := db.Stat()

			logger.Debug("Database connection pool status",
				"acquired", stats.AcquiredConns(),
				"idle", stats.IdleConns(),
				"total", stats.TotalConns(),
				"max", stats.MaxConns(),
			)

			time.Sleep(d)
		}
	}

	go monitor(time.Second * 10)
}

func initializeRouter(h *handler.ClientHandler) {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/clientes/:id/transacoes", h.CreateTransaction)
	r.GET("/clientes/:id/extrato", h.GetStatement)

	r.Use(middleware.TimeoutMiddleware(time.Second * 30))
	r.Run()
}
