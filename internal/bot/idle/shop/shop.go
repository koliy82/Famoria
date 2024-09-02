package shop

type Shop struct {
	Items map[string]SaleItem
}

func (s *Shop) AddItem(si SaleItem) {
	s.Items[si.item.Name] = si
}

func New() *Shop {
	return &Shop{
		Items: make(map[string]SaleItem),
	}
}
