package payment

import (
	"context"
	"famoria/internal/config"

	"github.com/mymmrac/telego"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var _ Repository = (*Mongo)(nil)

type Mongo struct {
	coll *mongo.Collection
	log  *zap.Logger
}

func (c *Mongo) Insert(m *telego.Message) error {
	p := m.SuccessfulPayment
	_, err := c.coll.InsertOne(context.TODO(), &Payment{
		OID:                     primitive.NewObjectID(),
		Currency:                p.Currency,
		TotalAmount:             p.TotalAmount,
		InvoicePayload:          p.InvoicePayload,
		ShippingOptionID:        p.ShippingOptionID,
		TelegramPaymentChargeID: p.TelegramPaymentChargeID,
		ProviderPaymentChargeID: p.ProviderPaymentChargeID,
		FromId:                  &m.From.ID,
	})
	if err != nil {
		c.log.Sugar().Error(err)
	}
	return nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("payments")
	return &Mongo{
		coll: coll,
		log:  log,
	}
}
