package domain

import (
	"errors"
	"testing"
)

func TestNewTransactionValid(t *testing.T) {
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

			assertNoError(t, err)
		})
	}
}

func TestNewTransactionInvalid(t *testing.T) {
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

			assertError(t, tt.expectedErr, err)
		})
	}
}

func assertError(t *testing.T, expected, got error) {
	t.Helper()

	if expected == nil && got != nil {
		t.Errorf("expected no error, but got: %v", got)
		return
	}

	if expected != nil && got == nil {
		t.Errorf("expected error: %v, but got none", expected)
		return
	}

	if !errors.Is(got, expected) {
		t.Errorf("expected error %v, but got %v", expected, got)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
