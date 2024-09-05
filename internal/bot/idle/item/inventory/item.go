package inventory

import (
	"famoria/internal/bot/idle/events"
	"famoria/internal/bot/idle/item/items"
	"famoria/internal/pkg/common"
	"famoria/internal/pkg/html"
	"fmt"
	"strconv"
)

type ShowItem struct {
	Name         items.Name
	Emoji        string
	CurrentLevel int
	MaxLevel     int
	Description  string
	Buffs        []events.Buff
}

func (si *ShowItem) FullDescription() string {
	header := html.Bold(si.Emoji+" "+si.Name.String()) + " (" + strconv.Itoa(si.CurrentLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.)" + "\n"
	body := si.Description + "\n"
	bDesc := ""
	for _, buff := range si.Buffs {
		bDesc += buff.Description() + "\n"
	}
	return header + body + html.CodeBlockWithLang(bDesc, "Усиления")
}

type ShopItem struct {
	Name        items.Name
	Emoji       string
	BuyLevel    int
	MaxLevel    int
	Description string
	Price       *common.Score
	Buffs       []events.Buff
}

func (si *ShopItem) FullDescription() string {
	header := html.Bold(si.Emoji+" "+si.Name.String()) + " (" + strconv.Itoa(si.BuyLevel) + "/" + strconv.Itoa(si.MaxLevel) + " ур.)" + "\n"
	price := "Цена: " + si.Price.GetFormattedScore() + " 💰 \n"
	body := si.Description + "\n"
	bDesc := ""
	for _, buff := range si.Buffs {
		bDesc += buff.Description() + "\n"
	}
	return header + price + html.Italic(body) + html.CodeBlockWithLang(bDesc, "Усиления")
}

func (si *ShopItem) SmallDescription() string {
	return fmt.Sprintf("%s - %s [%d/%d ур.] - %s", si.Emoji, si.Name.String(), si.BuyLevel, si.MaxLevel, si.Price.GetFormattedScore())
}