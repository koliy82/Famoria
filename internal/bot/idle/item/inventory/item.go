package inventory

import (
	"famoria/internal/bot/idle/event"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/html"
	"fmt"
	"strconv"
)

type ShowItem struct {
	ItemId       items.ItemId `bson:"name"`
	Emoji        string
	CurrentLevel int
	MaxLevel     int
	Description  string
	Buffs        []event.Buff
}

func (si *ShowItem) FullDescription() string {
	header := html.Bold(si.Emoji+" "+si.ItemId.String()) + " (" + strconv.Itoa(si.CurrentLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.)" + "\n"
	body := si.Description + "\n"
	bDesc := ""
	for _, buff := range si.Buffs {
		bDesc += buff.Description() + "\n"
	}
	return header + html.Italic(body) + html.CodeBlockWithLang(bDesc, "Усиления")
}

func (si *ShowItem) SmallDescription() string {
	return fmt.Sprintf("%s - %s [%d/%d ур.]", si.Emoji, si.ItemId.String(), si.CurrentLevel, si.MaxLevel)
}

type ShopItem struct {
	Name        items.ItemId
	Emoji       string
	BuyLevel    int
	MaxLevel    int
	Description string
	Price       *common.Score
	SalePrice   *common.Score
	Buffs       []event.Buff
}

func (si *ShopItem) FullDescription() string {
	header := html.Bold(si.Emoji+" "+si.Name.String()) + " (" + strconv.Itoa(si.BuyLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.)" + "\n"

	price := "Цена: "
	if si.SalePrice != nil {
		price += html.Strike(si.Price.GetFormattedScore()) + " " + si.SalePrice.GetFormattedScore()
	} else {
		price += si.Price.GetFormattedScore()
	}
	price += " 💰 \n"
	body := si.Description + "\n"
	bDesc := ""
	for _, buff := range si.Buffs {
		bDesc += buff.Description() + "\n"
	}
	return header + price + html.Italic(body) + html.CodeBlockWithLang(bDesc, "Усиления")
}

func (si *ShopItem) SmallDescription() string {
	if si.SalePrice != nil {
		return fmt.Sprintf("%s - %s [%d/%d ур.] - %s %s", si.Emoji, si.Name.String(), si.BuyLevel, si.MaxLevel, html.Strike(si.Price.GetFormattedScore()), si.SalePrice.GetFormattedScore())
	}
	return fmt.Sprintf("%s - %s [%d/%d ур.] - %s", si.Emoji, si.Name.String(), si.BuyLevel, si.MaxLevel, si.Price.GetFormattedScore())
}
