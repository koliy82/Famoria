package checkout

type Repository interface {
	Insert(ch *Checkout) error
}
