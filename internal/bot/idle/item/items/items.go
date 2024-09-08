package items

type Name int

const (
	MegaTap Name = iota
	TapCount
	TapPower
	GoldenDice
	InfiniteSpins
	LuckyClover
	MagicSpoon
	GrowthPotion
	EndlessMilk
	Subscribe
)

func (n Name) String() string {
	switch n {
	case MegaTap:
		return "Усиленный тап"
	case TapCount:
		return "Больше хомяков"
	case TapPower:
		return "Хомячий тренер"
	case GoldenDice:
		return "Золотые кости"
	case InfiniteSpins:
		return "Бесконечные крутки"
	case LuckyClover:
		return "Счастливый клевер"
	case MagicSpoon:
		return "Волшебная ложка"
	case GrowthPotion:
		return "Зелье роста"
	case EndlessMilk:
		return "Бесконечное молоко"
	case Subscribe:
		return "Сияющий кристалл"
	default:
		return "Неизвестный предмет"
	}
}
