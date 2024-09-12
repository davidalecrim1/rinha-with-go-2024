package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"rinha-with-go-2024/internal/domain"

	"github.com/gin-gonic/gin"
)

type ClientHandler struct {
	logger *slog.Logger
	svc    *domain.ClientService
}

func NewClientHandler(logger *slog.Logger, svc *domain.ClientService) *ClientHandler {
	return &ClientHandler{
		logger: logger,
		svc:    svc,
	}
}

// POST /clientes/:id/transacoes
func (h *ClientHandler) CreateTransaction(c *gin.Context) {
	clientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debug("invalid client id", "id", c.Param("id"), "error", err)
		c.Status(404)
		return
	}

	request := TransactionRequest{}
	if err := c.BindJSON(&request); err != nil {
		h.logger.Debug("invalid request body", "error", err)
		c.Status(422)
		return
	}

	t, err := domain.NewTransaction(
		clientID,
		request.Amount,
		request.Kind,
		request.Description,
	)
	if err != nil {
		h.logger.Debug("invalid transaction", "error", err)
		c.Status(422)
		return
	}

	client, err := h.svc.CreateTransaction(t)
	if errors.Is(err, domain.ErrClientDoesntExist) {
		h.logger.Debug("invalid client id", "id", clientID)
		c.Status(404)
		return
	}
	if err != nil {
		h.logger.Debug("the transaction was not perform correctly", "error", err)
		c.Status(422)
		return
	}

	response := TransactionResponse{
		Limit:   client.Limit,
		Balance: client.Balance,
	}
	c.JSON(200, response)
}

type TransactionRequest struct {
	Amount      int    `json:"valor"`
	Kind        string `json:"tipo"`
	Description string `json:"descricao"`
}

type TransactionResponse struct {
	Limit   int `json:"limite"`
	Balance int `json:"saldo"`
}
