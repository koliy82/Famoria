package inventory

//==============================

//type Attack struct {
//	atk int
//}
//
//func (a *Attack) Modifiers() []BuffModifier {
//	return []BuffModifier{
//		&AddTwoAttack{10},
//		&AddTwoAttack{20},
//	}
//}
//
//func (a *Attack) Apply(buffs ...BuffModifier) int {
//	for _, buff := range buffs {
//		buff.ForAttack(a)
//	}
//	return a.atk
//}

//==============================

//type BuffModifier interface {
//	ForAttack(attack *Attack)
//}

//type MinusAttack struct{}
//
//func (m *MinusAttack) ForAttack(attack *Attack) {
//	attack.atk -= 1
//}
//
//type AddTwoAttack struct {
//	buffAtk int
//}
//
//func (a *AddTwoAttack) ForAttack(attack *Attack) {
//	attack.atk += a.buffAtk
//}

//==============================

type Inventory struct {
	Items []Item
}
