package shop

import (
	"famoria/internal/bot/idle/item"
)

type Shop struct {
	Items []item.Item
}

func (s *Shop) AddItem(si item.Item) {
	s.Items = append(s.Items, si)
}

func New() *Shop {
	return &Shop{}
}
