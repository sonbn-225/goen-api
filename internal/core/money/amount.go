package money

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type Amount struct {
	decimal.Decimal
}

func Zero() Amount {
	return Amount{Decimal: decimal.Zero}
}

func NewFromString(v string) (Amount, error) {
	d, err := decimal.NewFromString(strings.TrimSpace(v))
	if err != nil {
		return Amount{}, err
	}
	return Amount{Decimal: d}, nil
}

func MustFromString(v string) Amount {
	a, err := NewFromString(v)
	if err != nil {
		panic(err)
	}
	return a
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Decimal.String())
}

func (a *Amount) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		a.Decimal = decimal.Zero
		return nil
	}

	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return err
		}
		d, err := decimal.NewFromString(strings.TrimSpace(s))
		if err != nil {
			return fmt.Errorf("invalid decimal string: %w", err)
		}
		a.Decimal = d
		return nil
	}

	d, err := decimal.NewFromString(string(trimmed))
	if err != nil {
		return fmt.Errorf("invalid decimal number: %w", err)
	}
	a.Decimal = d
	return nil
}
