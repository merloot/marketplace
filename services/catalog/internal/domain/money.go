package domain

type Money struct {
	amount   int64
	currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}

	if amount < 0 {
		return Money{}, ErrNegativePrice
	}

	return Money{
		amount:   amount,
		currency: currency,
	}, nil
}

func (m Money) IsNegative() bool { return m.amount < 0 }
func (m Money) Amount() int64    { return m.amount }
func (m Money) Currency() string { return m.currency }
