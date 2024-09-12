package repository

import (
	"context"
	"log/slog"

	"rinha-with-go-2024/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientRepository struct {
	logger *slog.Logger
	db     *pgxpool.Pool
	tx     pgx.Tx
}

func NewClientRepository(logger *slog.Logger,
	db *pgxpool.Pool,
) *ClientRepository {
	return &ClientRepository{logger: logger, db: db}
}

// begin transaction
// commit and rollbackas

func (r *ClientRepository) Begin(ctx context.Context) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	r.tx = tx
	return nil
}

func (r *ClientRepository) Commit(ctx context.Context) error {
	return r.tx.Commit(ctx)
}

func (r *ClientRepository) Rollback(ctx context.Context) error {
	return r.tx.Rollback(ctx)
}

func (r *ClientRepository) ExecuteTransaction(ctx context.Context, t *domain.Transaction) error {
	query := `
	INSERT INTO transactions (clientId, amount, kind, description) 
	VALUES ($1, $2, $3, $4);
	`
	_, err := r.tx.Exec(ctx, query, t.ClientID, t.Amount, t.Kind, t.Description)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23503" {
			return domain.ErrClientDoesntExist
		}
	}

	return err
}

func (r *ClientRepository) UpdateClientBalance(ctx context.Context, clientID int, amount int) error {
	err := r.getClientBalanceForUpdate(ctx, clientID)
	if err != nil {
		return err
	}

	// This query ensures the balance is not updated if it
	// will be below the client's limit (like a credit in the bank).
	query := `
	UPDATE clients
	SET balance = balance + ($1),
		UpdatedAt = NOW()
	WHERE id = $2
	AND limitBalance + (balance + ($1)) > 0;
	`
	result, err := r.tx.Exec(ctx, query, amount, clientID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTransactionOverClientLimit
	}

	return nil
}

func (r *ClientRepository) getClientBalanceForUpdate(ctx context.Context, clientID int) error {
	row := r.tx.QueryRow(ctx, `SELECT balance FROM clients WHERE id = $1 FOR UPDATE`, clientID)
	var balance int
	if err := row.Scan(&balance); err != nil {
		return err
	}
	return nil
}

func (r *ClientRepository) GetClientBalance(ctx context.Context, clientID int) (*domain.Client, error) {
	var client *domain.Client = &domain.Client{ID: clientID}
	query := `
	SELECT limitBalance, balance, UpdatedAt
	FROM clients
	WHERE id = $1;
	`
	err := r.db.QueryRow(ctx, query, clientID).Scan(&client.Limit, &client.Balance, &client.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrClientDoesntExist
	}
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *ClientRepository) GetClientTransactions(ctx context.Context, clientID int) (*[]domain.Transaction, error) {
	query := `
	SELECT amount, kind, description, updatedat
	FROM public.transactions
	WHERE transactions.clientId = $1
	ORDER BY UpdatedAt DESC
	LIMIT 10;
	`

	rows, err := r.db.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}

	return r.mapTransactions(rows)
}

func (r *ClientRepository) mapTransactions(rows pgx.Rows) (*[]domain.Transaction, error) {
	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		err := rows.Scan(&t.Amount, &t.Kind, &t.Description, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, t)
	}

	return &transactions, nil
}
