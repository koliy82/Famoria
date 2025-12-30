package inventory

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/items"
)

type Item struct {
	Id           items.ItemId `bson:"name"`
	CurrentLevel int
}

func (i *Item) GetItem(manager *item.Manager) *item.Item {
	return manager.GetItem(i.Id)
}

func (i *Item) GetBuffs(manager *item.Manager) []event.Buff {
	buffs := i.GetItem(manager).Buffs[i.CurrentLevel]
	if buffs == nil {
		manager.Log.Sugar().Error("Buffs not found", i.Id, i.CurrentLevel)
	}
	return buffs
}

type Inventory struct {
	Items map[items.ItemId]Item `bson:"items"`
}

//func (i *Inventory) GetItems(manager *item.Manager) []*ShowItem {
//	list := make([]*ShowItem, 0, len(i.Items))
//	for _, i := range i.Items {
//		mi := i.GetItem(manager)
//		list = append(list, &ShowItem{
//			ItemId:       mi.ItemId,
//			CurrentLevel: i.CurrentLevel,
//			MaxLevel:     mi.MaxLevel,
//			Description:  mi.Description,
//			Buffs:        mi.Buffs[i.CurrentLevel],
//		})
//	}
//	return list
//}
