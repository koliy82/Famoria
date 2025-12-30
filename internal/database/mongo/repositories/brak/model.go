package brak

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/event/events"
	"famoria/internal/bot/idle/item"
	"famoria/internal/bot/idle/item/inventory"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/database/mongo/repositories/user"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/plural"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Brak struct {
	OID            primitive.ObjectID   `bson:"_id"`
	FirstUserID    int64                `bson:"first_user_id"`
	SecondUserID   int64                `bson:"second_user_id"`
	ChatID         int64                `bson:"chat_id,omitempty"`
	CreateDate     time.Time            `bson:"create_date"`
	BabyUserID     *int64               `bson:"baby_user_id"`
	BabyCreateDate *time.Time           `bson:"baby_create_date"`
	Score          *common.Score        `bson:"score"`
	SubscribeEnd   *time.Time           `bson:"subscribe_end"`
	Inventory      *inventory.Inventory `bson:"inventory"`
	Events         *events.Events       `bson:"events"`
}

func (b *Brak) ApplyBuffs(manager *item.Manager) {
	if b.Inventory == nil {
		return
	}
	si, ok := b.Inventory.Items[items.Subscribe]
	if b.IsSub() {
		if !ok {
			si = inventory.Item{
				Id:           items.Subscribe,
				CurrentLevel: 0,
			}
			b.Inventory.Items[items.Subscribe] = si
		}
	} else {
		if ok {
			delete(b.Inventory.Items, items.Subscribe)
		}
	}
	b.Events.DefaultStats()
	b.Events.Shop = &event.Base{}
	for _, i := range b.Inventory.Items {
		for _, buff := range i.GetBuffs(manager) {
			switch buff.Type() {
			case event.Hamster:
				buff.Apply(&b.Events.Hamster.Base)
			case event.Casino:
				buff.Apply(&b.Events.Casino.Base)
			case event.GrowKid:
				buff.Apply(&b.Events.GrowKid.Base)
			case event.Mining:
				buff.Apply(&b.Events.Mining.Base)
			case event.Shop:
				buff.Apply(b.Events.Shop)
			case event.Subscribe:
				continue
			}
		}
	}
}

func (b *Brak) IsSub() bool {
	return b.SubscribeEnd != nil && time.Now().Before(*b.SubscribeEnd)
}

func (b *Brak) SubDaysCount() int {
	if b.SubscribeEnd == nil {
		return 0
	}
	return int(b.SubscribeEnd.Sub(time.Now()).Hours() / 24)
}

func (b *Brak) AddSubDays(d time.Duration) {
	if b.SubscribeEnd == nil {
		subEnd := time.Now().Add(d)
		b.SubscribeEnd = &subEnd
	} else {
		subEnd := b.SubscribeEnd.Add(d)
		b.SubscribeEnd = &subEnd
	}
}

type UsersBrak struct {
	Brak   *Brak
	First  *user.User
	Second *user.User
	Baby   *user.User
}

// PartnerID returns the partner's ID by the user's ID
func (b *Brak) PartnerID(userID int64) int64 {
	if b.FirstUserID == userID {
		return b.SecondUserID
	}
	return b.FirstUserID
}

// Duration returns the duration of the relationship
func (b *Brak) Duration() string {
	duration := time.Now().Sub(b.CreateDate)
	hours := int(duration.Hours())

	if hours < 1 {
		return "молодожены"
	}

	days := hours / 24
	months := days / 30
	years := days / 365

	if years > 0 {
		return fmt.Sprintf("%d %s", years, plural.Declension(years, "год", "года", "лет"))
	}

	if months > 0 {
		return fmt.Sprintf("%d %s", months, plural.Declension(months, "месяц", "месяца", "месяцев"))
	}

	if days > 0 {
		return fmt.Sprintf("%d %s", days, plural.Declension(days, "день", "дня", "дней"))
	}

	return fmt.Sprintf("%d %s", hours, plural.Declension(hours, "час", "часа", "часов"))
}

func (b *Brak) DurationKid() string {
	if b.BabyCreateDate == nil {
		return ""
	}
	duration := time.Now().Sub(*b.BabyCreateDate)
	seconds := int(duration.Seconds())

	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	months := days / 30
	years := days / 365

	switch {
	case years > 0:
		return fmt.Sprintf("%d %s", years, plural.Declension(years, "год", "года", "лет"))
	case months > 0:
		return fmt.Sprintf("%d %s", months, plural.Declension(months, "месяц", "месяца", "месяцев"))
	case days > 0:
		return fmt.Sprintf("%d %s", days, plural.Declension(days, "день", "дня", "дней"))
	case hours > 0:
		return fmt.Sprintf("%d %s", hours, plural.Declension(hours, "час", "часа", "часов"))
	case minutes > 0:
		return fmt.Sprintf("%d %s", minutes, plural.Declension(minutes, "минута", "минуты", "минут"))
	default:
		return fmt.Sprintf("%d %s", seconds, plural.Declension(seconds, "секунда", "секунды", "секунд"))
	}
}

func (b *Brak) GetAvailableItems(manager *item.Manager) []*inventory.ShopItem {
	list := make([]*inventory.ShopItem, 0, len(manager.Items))
	for _, mi := range manager.Items {
		if mi.MaxLevel == 0 {
			continue
		}
		current, ok := b.Inventory.Items[mi.ItemId]
		if ok == false {
			si := &inventory.ShopItem{
				Name:        mi.ItemId,
				Emoji:       mi.Emoji,
				BuyLevel:    1,
				MaxLevel:    mi.MaxLevel,
				Description: mi.Description,
				Price:       mi.Prices[1],
				Buffs:       mi.Buffs[1],
			}
			if b.Events.Shop.Sale > 0 {
				si.SalePrice = si.Price.GetSaleScore(b.Events.Shop.Sale)
			}
			list = append(list, si)
			continue
		}
		if current.CurrentLevel >= mi.MaxLevel {
			continue
		}
		si := &inventory.ShopItem{
			Name:        mi.ItemId,
			Emoji:       mi.Emoji,
			BuyLevel:    current.CurrentLevel + 1,
			MaxLevel:    mi.MaxLevel,
			Description: mi.Description,
			Price:       mi.Prices[current.CurrentLevel+1],
			Buffs:       mi.Buffs[current.CurrentLevel+1],
		}
		if b.Events.Shop.Sale > 0 {
			si.SalePrice = si.Price.GetSaleScore(b.Events.Shop.Sale)
		}
		list = append(list, si)
	}
	return list
}
