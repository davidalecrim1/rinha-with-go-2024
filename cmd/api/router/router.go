package router

import (
	"log/slog"
	"rinha-with-go-2024/cmd/api/handler"
	"rinha-with-go-2024/internal/domain"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(logger *slog.Logger, r *gin.Engine, svc *domain.ClientService) {
	h := handler.NewClientHandler(logger, svc)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/clientes/:id/transacoes", h.CreateTransaction)
	r.GET("/clientes/:id/extrato", h.GetStatement)
}
