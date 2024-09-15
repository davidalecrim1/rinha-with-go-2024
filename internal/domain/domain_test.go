package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_New_Valid(t *testing.T) {
	tests := []struct {
		name  string
		given Transaction
	}{
		{
			name: "valid credit transaction",
			given: Transaction{
				ClientID:    10,
				Amount:      1000,
				Kind:        "c",
				Description: "test",
			},
		},
		{
			name: "valid debit transaction",
			given: Transaction{
				ClientID:    10,
				Amount:      500,
				Kind:        "d",
				Description: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransaction(
				tt.given.ClientID,
				tt.given.Amount,
				tt.given.Kind,
				tt.given.Description)

			assert.NoError(t, err)
		})
	}
}

func TestTransaction_New_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		given       Transaction
		expectedErr error
	}{
		{
			name: "invalid transaction description",
			given: Transaction{
				ClientID:    10,
				Amount:      1000,
				Kind:        "c",
				Description: "description greater then 10 characters",
			},
			expectedErr: ErrInvalidTransaction,
		},
		{
			name: "invalid transaction kind",
			given: Transaction{
				ClientID:    10,
				Amount:      1000,
				Kind:        "invalid kind",
				Description: "descrip",
			},
			expectedErr: ErrInvalidTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTransaction(
				tt.given.ClientID,
				tt.given.Amount,
				tt.given.Kind,
				tt.given.Description)

			assert.Error(t, tt.expectedErr, err)
		})
	}
}

func TestClient_New_Valid(t *testing.T) {
	updatedAt := time.Now()

	tests := []struct {
		name  string
		given Client
	}{
		{
			name: "valid client",
			given: Client{
				ID:        1,
				Limit:     1000,
				Balance:   500,
				UpdatedAt: updatedAt,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(
				tt.given.ID,
				tt.given.Limit,
				tt.given.Balance,
				tt.given.UpdatedAt)

			assert.Equal(t, tt.given.ID, client.ID)
			assert.Equal(t, tt.given.Limit, client.Limit)
			assert.Equal(t, tt.given.Balance, client.Balance)
			assert.Equal(t, tt.given.UpdatedAt, client.UpdatedAt)
		})
	}
}
