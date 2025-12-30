package common

import (
	"encoding/json"
)

type ProviderData struct {
	Receipt Receipt `json:"receipt"`
}

func (d *ProviderData) ToJson() (string, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type Receipt struct {
	Items []Item  `json:"items"`
	Email *string `json:"email"`
}

type Item struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Amount      Amount `json:"amount"`
	VatCode     int    `json:"vat_code"`
}

type Amount struct {
	Currency string `json:"currency"`
	Value    string `json:"value"`
}
