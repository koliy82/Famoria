package inventory

import (
	"famoria/internal/bot/idle/event"
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

func (i *Item) GetBuffs(manager *item.Manager) []event.Buff {
	buffs := i.GetItem(manager).Buffs[i.CurrentLevel]
	if buffs == nil {
		manager.Log.Sugar().Error("Buffs not found", i.Name, i.CurrentLevel)
	}
	return buffs
}

type Inventory struct {
	Items map[items.Name]Item `bson:"items"`
	Base  event.Base          `bson:"-"`
}

func (i *Inventory) GetItems(manager *item.Manager) []*ShowItem {
	list := make([]*ShowItem, 0, len(i.Items))
	for _, i := range i.Items {
		mi := i.GetItem(manager)
		list = append(list, &ShowItem{
			Name:         mi.Name,
			CurrentLevel: i.CurrentLevel,
			MaxLevel:     mi.MaxLevel,
			Description:  mi.Description,
			Buffs:        mi.Buffs[i.CurrentLevel],
		})
	}
	return list
}

func (i *Inventory) GetAvailableItems(manager *item.Manager) []*ShopItem {
	list := make([]*ShopItem, 0, len(manager.Items))
	for _, mi := range manager.Items {
		if mi.MaxLevel == 0 {
			continue
		}
		current, ok := i.Items[mi.Name]
		if ok == false {
			list = append(list, &ShopItem{
				Name:        mi.Name,
				Emoji:       mi.Emoji,
				BuyLevel:    1,
				MaxLevel:    mi.MaxLevel,
				Description: mi.Description,
				Price:       mi.Prices[1],
				Buffs:       mi.Buffs[1],
			})
			continue
		}
		if current.CurrentLevel >= mi.MaxLevel {
			println("current.CurrentLevel >= mi.MaxLevel")
			continue
		}
		list = append(list, &ShopItem{
			Name:        mi.Name,
			Emoji:       mi.Emoji,
			BuyLevel:    current.CurrentLevel + 1,
			MaxLevel:    mi.MaxLevel,
			Description: mi.Description,
			Price:       mi.Prices[current.CurrentLevel+1],
			Buffs:       mi.Buffs[current.CurrentLevel+1],
		})
	}
	return list
}
