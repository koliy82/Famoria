package items

type Name int

const (
	MegaTap Name = iota
	TapCount
	TapPower
)

func (n Name) String() string {
	switch n {
	case MegaTap:
		return "Усиленный тап"
	case TapCount:
		return "Больше хомяков"
	case TapPower:
		return "Хомячий тренер"
	default:
		return "Неизвестный предмет"
	}
}
