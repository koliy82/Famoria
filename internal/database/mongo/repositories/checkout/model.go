package checkout

import (
	"famoria/internal/database/mongo/repositories/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Checkout struct {
	OID              primitive.ObjectID `bson:"_id"`
	ID               string             `bson:"id"`
	FromId           int64              `bson:"from_id"`
	From             *user.User         `json:"from"`
	Currency         string             `bson:"currency"`
	TotalAmount      int                `bson:"total_amount"`
	InvoicePayload   string             `bson:"invoice_payload"`
	ShippingOptionID *string            `bson:"shipping_option_id"`
}
