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

func NewClient(id int, limit int, balance int) *Client {
	return &Client{
		ID:      id,
		Limit:   limit,
		Balance: balance,
	}
}

type Transaction struct {
	TransactionID int
	ClientID      int
	Amount        int
	Kind          string
	Description   string
	UpdatedAt     time.Time
}

func NewTransaction(
	clientId int,
	amount int,
	kind string,
	description string,
) (*Transaction, error) {
	t := &Transaction{
		ClientID:    clientId,
		Amount:      amount,
		Kind:        kind,
		Description: description,
	}

	// TODO: Do i need to check if amount is 0?

	if err := t.validKind(); err != nil {
		return nil, err
	}

	if err := t.validDescription(); err != nil {
		return nil, err
	}

	if kind == "d" {
		t.Amount = -t.Amount
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
	UpdateClientBalance(ctx context.Context, clientID int, transactionAmount int) error
	ExecuteTransaction(ctx context.Context, t *Transaction) error
	GetClientBalance(ctx context.Context, clientID int) (*Client, error)
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func (s *ClientService) CreateTransaction(t *Transaction) (*Client, error) {
	ctx := context.Background()
	var err error

	defer func() {
		if err != nil {
			s.logger.Error("rolling back transaction because of error", "error", err)
			s.repo.Rollback(ctx)
		}
	}()

	if err := s.repo.Begin(ctx); err != nil {
		s.logger.Error("failed to begin transaction", "error", err)
		return nil, err
	}

	if err := s.repo.ExecuteTransaction(ctx, t); err != nil {
		s.logger.Error("failed to execute transaction", "error", err)
		return nil, err
	}

	if err := s.repo.UpdateClientBalance(ctx, t.ClientID, t.Amount); err != nil {
		s.logger.Error("failed to update client balance", "error", err)
		return nil, err
	}

	if err := s.repo.Commit(ctx); err != nil {
		s.logger.Error("failed to commit transaction", "error", err)
		return nil, err
	}

	client, err := s.repo.GetClientBalance(context.Background(), t.ClientID)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *ClientService) GetTransactions(clientId int) (*[]Transaction, error) {
	return nil, nil // TODO
}
