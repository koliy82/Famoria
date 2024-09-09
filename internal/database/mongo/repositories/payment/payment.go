package payment

import "github.com/mymmrac/telego"

type Repository interface {
	Insert(p *telego.Message) error
}
