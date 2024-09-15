//go:build integration

package repository

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"rinha-with-go-2024/config/env"
	"rinha-with-go-2024/internal/domain"
	"strconv"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestClientRepository_New(t *testing.T) {
	db := initializeDatabase(t)
	logger := initializeLogger()
	defer db.Close()

	t.Run("should return a new repository", func(t *testing.T) {
		repo := NewClientRepository(logger, db)
		assert.NotNil(t, repo)
	})
}

func TestClientRepository_ExecuteTransaction(t *testing.T) {
	db := initializeDatabase(t)
	logger := initializeLogger()
	defer db.Close()

	repo := NewClientRepository(logger, db)

	t.Run("valid credit transaction", func(t *testing.T) {
		clientId := 1
		amount := 1000

		transaction, err := domain.NewTransaction(
			clientId,
			uint(amount),
			"c",
			"descricao",
		)
		assert.NoError(t, err)
		err = repo.ExecuteTransaction(context.Background(), transaction)
		assert.NoError(t, err)

		client := domain.Client{}
		err = db.QueryRow(context.Background(), "SELECT * FROM clients WHERE id = $1", clientId).Scan(&client.ID, &client.Limit, &client.Balance, &client.UpdatedAt)

		assert.NoError(t, err)
		assert.Equal(t, amount, client.Balance)
		assert.Equal(t, clientId, client.ID)

		t.Cleanup(cleanUpClientRepository(t, db, clientId))
	})

	t.Run("valid credit transaction to unexisting client", func(t *testing.T) {
		clientId := 10000
		amount := 1000

		transaction, err := domain.NewTransaction(
			clientId,
			uint(amount),
			"c",
			"descricao",
		)
		assert.NoError(t, err)
		err = repo.ExecuteTransaction(context.Background(), transaction)
		assert.Error(t, err, domain.ErrClientDoesntExist)
	})

	t.Run("valid debit transaction within limit", func(t *testing.T) {
		clientId := 1
		amount := 1000

		transaction, err := domain.NewTransaction(
			clientId,
			uint(amount),
			"d",
			"descricao",
		)

		assert.NoError(t, err)
		err = repo.ExecuteTransaction(context.Background(), transaction)
		assert.NoError(t, err)

		client := domain.Client{}
		err = db.QueryRow(context.Background(), "SELECT * FROM clients WHERE id = $1", clientId).Scan(&client.ID, &client.Limit, &client.Balance, &client.UpdatedAt)
		assert.NoError(t, err)
		amountAsNegative := -amount

		assert.Equal(t, amountAsNegative, client.Balance)
		assert.Equal(t, clientId, client.ID)

		t.Cleanup(cleanUpClientRepository(t, db, clientId))
	})

	t.Run("valid debit transaction over the limit", func(t *testing.T) {
		clientId := 1
		amount := 1000000000

		transaction, err := domain.NewTransaction(
			clientId,
			uint(amount),
			"d",
			"descricao",
		)

		assert.NoError(t, err)
		err = repo.ExecuteTransaction(context.Background(), transaction)
		assert.ErrorIs(t, err, domain.ErrTransactionOverClientLimit)
	})

	t.Run("concorrent debit updates without race condition", func(t *testing.T) {
		clientId := 2
		amount := 5000
		amountAsNegative := -amount
		concorrentUpdates := 10

		transaction, err := domain.NewTransaction(
			clientId,
			uint(amount),
			"d",
			"descricao",
		)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < concorrentUpdates; i++ {
			wg.Add(1)

			go func(t *testing.T, transaction domain.Transaction) {
				err := repo.ExecuteTransaction(context.Background(), &transaction)
				assert.NoError(t, err)
				defer wg.Done()
			}(t, *transaction)
		}

		wg.Wait()

		client := domain.Client{}
		err = db.QueryRow(context.Background(), "SELECT * FROM clients WHERE id = $1", clientId).Scan(&client.ID, &client.Limit, &client.Balance, &client.UpdatedAt)
		assert.NoError(t, err)

		assert.Equal(t, amountAsNegative*concorrentUpdates, client.Balance)
		t.Cleanup(cleanUpClientRepository(t, db, clientId))
	})
}

func TestClientRepository_GetClientBalance(t *testing.T) {
	db := initializeDatabase(t)
	logger := initializeLogger()
	defer db.Close()

	repo := NewClientRepository(logger, db)

	t.Run("valid client check balance", func(t *testing.T) {
		clientId := 1
		expectedBalance := 0

		client, err := repo.GetClientBalance(context.Background(), clientId)
		assert.NoError(t, err)
		assert.Equal(t, expectedBalance, client.Balance)
	})

	t.Run("invalid client check balance", func(t *testing.T) {
		clientId := 100 // doesnt exist in initialized db

		client, err := repo.GetClientBalance(context.Background(), clientId)
		assert.ErrorIs(t, err, domain.ErrClientDoesntExist)
		assert.Nil(t, client)
	})

}

func TestClientRepository_GetClientTransactions(t *testing.T) {
	db := initializeDatabase(t)
	logger := initializeLogger()
	defer db.Close()

	repo := NewClientRepository(logger, db)

	t.Run("get client existing transactions", func(t *testing.T) {
		clientId := 1

		transaction := &domain.Transaction{
			ClientID:    clientId,
			Amount:      1000,
			Kind:        "c",
			Description: "descricao",
		}
		createTransactions(t, 10, db, transaction)

		tt, err := repo.GetClientTransactions(context.Background(), clientId)
		assert.NoError(t, err)
		assert.Len(t, *tt, 10)

		for _, transc := range *tt {
			assert.Equal(t, transc.Kind, transaction.Kind)
			assert.Equal(t, transc.Amount, transaction.Amount)
			assert.Equal(t, transc.Description, transaction.Description)
		}

		t.Cleanup(cleanUpClientRepository(t, db, clientId))
	})

	t.Run("get client transaction without any existing", func(t *testing.T) {
		clientId := 2

		tt, err := repo.GetClientTransactions(context.Background(), clientId)
		assert.NoError(t, err)
		assert.Len(t, *tt, 0)
	})

}

func createTransactions(t *testing.T, quantity int, db *pgxpool.Pool, transaction *domain.Transaction) {
	t.Helper()

	for i := 0; i < quantity; i++ {

		_, err := db.Exec(context.Background(), "INSERT INTO transactions (clientId, amount, kind, description) VALUES ($1, $2, $3, $4)", transaction.ClientID, transaction.Amount, transaction.Kind, transaction.Description)
		assert.NoError(t, err)
	}
}

func initializeDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

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

func initializeLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

func cleanUpClientRepository(t *testing.T, db *pgxpool.Pool, clientId int) func() {
	return func() {
		_, err := db.Exec(context.Background(), "DELETE FROM transactions WHERE clientId = $1", clientId)
		assert.NoError(t, err)

		_, err = db.Exec(context.Background(), "UPDATE clients SET balance = 0 WHERE id = $1", clientId)
		assert.NoError(t, err)
	}
}
