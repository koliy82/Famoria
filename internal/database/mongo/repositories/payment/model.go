package payment

import "go.mongodb.org/mongo-driver/bson/primitive"

type Payment struct {
	OID                     primitive.ObjectID `bson:"_id"`
	Currency                string             `bson:"currency"`
	TotalAmount             int                `bson:"total_amount"`
	InvoicePayload          string             `bson:"invoice_payload"`
	ShippingOptionID        string             `bson:"shipping_option_id"`
	TelegramPaymentChargeID string             `bson:"telegram_payment_charge_id"`
	ProviderPaymentChargeID string             `bson:"provider_payment_charge_id"`
	FromId                  *int64             `bson:"from_id"`
}
