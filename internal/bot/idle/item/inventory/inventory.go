package inventory

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/items"
)

type Item struct {
	Name         items.Name
	CurrentLevel int
}

func (i *Item) GetItem(manager *item.Manager) *item.Item {
	return manager.GetItem(i.Name)
}

func (i *Item) GetBuffs(manager *item.Manager) []events.Buff {
	buffs := i.GetItem(manager).Buffs[i.CurrentLevel]
	if buffs == nil {
		manager.Log.Sugar().Error("Buffs not found", i.Name, i.CurrentLevel)
	}
	return buffs
}

type Inventory struct {
	Items []Item
}
