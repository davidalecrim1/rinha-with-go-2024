package domain

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var (
	ErrInvalidTransaction         = errors.New("invalid transaction")
	ErrTransactionOverClientLimit = errors.New("transaction over the client's limit")
	ErrClientDoesntExist          = errors.New("client doesn't exist")
)

type Client struct {
	ID        int
	Limit     int
	Balance   int
	UpdatedAt time.Time
}

func NewClient(id int, limit int, balance int, updatedAt time.Time) *Client {
	return &Client{
		ID:        id,
		Limit:     limit,
		Balance:   balance,
		UpdatedAt: updatedAt,
	}
}

type Transaction struct {
	TransactionID int
	ClientID      int
	Amount        uint
	Kind          string
	Description   string
	UpdatedAt     time.Time
}

func NewTransaction(
	clientId int,
	amount uint,
	kind string,
	description string,
) (*Transaction, error) {
	t := &Transaction{
		ClientID:    clientId,
		Amount:      amount,
		Kind:        kind,
		Description: description,
	}

	if err := t.validKind(); err != nil {
		return nil, err
	}

	if err := t.validDescription(); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Transaction) validKind() error {
	if t.Kind == "c" || t.Kind == "d" {
		return nil
	}

	return ErrInvalidTransaction
}

func (t *Transaction) validDescription() error {
	if len(t.Description) > 0 && len(t.Description) <= 10 {
		return nil
	}

	return ErrInvalidTransaction
}

type ClientService struct {
	logger *slog.Logger
	repo   ClientRepository
}

func NewClientRepository(logger *slog.Logger, repo ClientRepository) *ClientService {
	return &ClientService{
		logger: logger,
		repo:   repo,
	}
}

type ClientRepository interface {
	ExecuteTransaction(ctx context.Context, t *Transaction) error
	GetClientBalance(ctx context.Context, clientID int) (*Client, error)
	GetClientTransactions(ctx context.Context, clientID int) (*[]Transaction, error)
}

func (s *ClientService) CreateTransaction(ctx context.Context, t *Transaction) (*Client, error) {
	err := s.repo.ExecuteTransaction(ctx, t)
	if err != nil {
		s.logger.Error("failed to execute transaction", "error", err)
		return nil, err
	}

	return s.repo.GetClientBalance(ctx, t.ClientID)
}

func (s *ClientService) GetStatement(ctx context.Context, clientId int) (*Client, *[]Transaction, error) {
	client, err := s.repo.GetClientBalance(ctx, clientId)
	if err != nil {
		return nil, nil, err
	}

	transactions, err := s.repo.GetClientTransactions(ctx, clientId)
	if err != nil {
		return nil, nil, err
	}

	return client, transactions, nil
}
