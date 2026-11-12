package domain_test

import (
	"testing"

	"github.com/merloot/marketplace/services/catalog/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		amount      int64
		currency    string
		expectedErr error
	}{
		{
			name:        "valid money",
			amount:      10050,
			currency:    "RUB",
			expectedErr: nil,
		},
		{
			name:        "negative amount",
			amount:      -10,
			currency:    "RUB",
			expectedErr: domain.ErrNegativePrice,
		},
		{
			name:        "empty currency",
			amount:      100,
			currency:    "",
			expectedErr: domain.ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := domain.NewMoney(tt.amount, tt.currency)
			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.amount, m.Amount())
				assert.Equal(t, tt.currency, m.Currency())
			}
		})
	}
}
