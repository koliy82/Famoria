package item

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/item/items"
	"go.uber.org/zap"
)

type Item struct {
	Name        items.Name
	MaxLevel    int
	Buffs       map[int][]events.Buff
	Description string
	Prices      map[int]uint64
}

type Manager struct {
	Log   *zap.Logger
	items map[items.Name]*Item
}

func (i *Manager) GetItem(name items.Name) *Item {
	item := i.items[name]
	if item == nil {
		i.Log.Sugar().Error("Item not found", name)
	}
	return item
}

func New(log *zap.Logger) *Manager {
	return &Manager{
		Log: log,
		items: map[items.Name]*Item{
			items.KitStart: {
				Name:     items.KitStart,
				MaxLevel: 1,
				Buffs: map[int][]events.Buff{
					1: {},
				},
				Prices: map[int]uint64{
					1: 250,
					2: 500,
					3: 1000,
					4: 2500,
					5: 5000,
				},
			},
		},
	}
}
