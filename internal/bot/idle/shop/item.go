package shop

import "famoria/internal/bot/idle/inventory"

type SaleItem struct {
	item          inventory.Item
	Description   string `default:"Item description"`
	BasePrice     uint64 `default:"0"`
	PriceIncrease uint64 `default:"0"`
}
