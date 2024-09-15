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
	ctx := c.Request.Context()
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

	client, err := h.svc.CreateTransaction(ctx, t)
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
	Amount      uint   `json:"valor"`
	Kind        string `json:"tipo"`
	Description string `json:"descricao"`
}

type TransactionResponse struct {
	Limit   int `json:"limite"`
	Balance int `json:"saldo"`
}

// GET /clientes/:id/extrato
func (h *ClientHandler) GetStatement(c *gin.Context) {
	ctx := c.Request.Context()
	clientID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.logger.Debug("invalid client id", "id", c.Param("id"), "error", err)
		c.Status(404)
		return
	}

	client, transactions, err := h.svc.GetStatement(ctx, clientID)
	if errors.Is(err, domain.ErrClientDoesntExist) {
		h.logger.Debug("invalid client id", "id", clientID)
		c.Status(404)
		return
	}
	if err != nil {
		h.logger.Error("error getting statement, maybe because of concorrent updates", "error", err)
		c.Status(422)
		return
	}

	transactionsResponse := make([]TransactionStatementResponse, 0, len(transactions))
	for _, t := range transactions {
		transactionsResponse = append(transactionsResponse, TransactionStatementResponse{
			Amount:      t.Amount,
			Kind:        t.Kind,
			Description: t.Description,
			UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05.000000Z"),
		})
	}

	response := &StatementResponse{
		Balance: StatementBalanceResponse{
			Total:       client.Balance,
			StatementAt: client.UpdatedAt.Format("2006-01-02T15:04:05.000000Z"),
			Limit:       client.Limit,
		},
		Transactions: transactionsResponse,
	}

	c.JSON(200, response)
}

type StatementResponse struct {
	Balance      StatementBalanceResponse       `json:"saldo"`
	Transactions []TransactionStatementResponse `json:"ultimas_transacoes"`
}

type StatementBalanceResponse struct {
	Total       int    `json:"total"`
	StatementAt string `json:"data_extrato"`
	Limit       int    `json:"limite"`
}

type TransactionStatementResponse struct {
	Amount      uint   `json:"valor"`
	Kind        string `json:"tipo"`
	Description string `json:"descricao"`
	UpdatedAt   string `json:"realizada_em"`
}
