package admin

import (
	"context"
	"famoria/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Mongo struct {
	coll   *mongo.Collection
	Admins []*Admin `bson:"admins"`
	log    *zap.Logger
}

func (m *Mongo) ActualData() {
	var result []*Admin
	cursor, err := m.coll.Find(context.TODO(), bson.D{})
	if err != nil {
		m.log.Sugar().Error(err)
		return
	}
	err = cursor.All(context.Background(), &result)
	if err != nil {
		m.log.Sugar().Error(err)
		return
	}
	m.Admins = result
}

func (m *Mongo) Add(admin *Admin) {
	m.Admins = append(m.Admins, admin)
	_, err := m.coll.InsertOne(context.TODO(), admin)
	if err != nil {
		m.log.Sugar().Error(err)
		return
	}
	m.log.Sugar().Info("Admin added: ", admin.UserID)
}

func (m *Mongo) Remove(userId int64) {
	_, err := m.coll.DeleteOne(context.TODO(), bson.M{"user_id": userId})
	if err != nil {
		m.log.Sugar().Error(err)
		return
	}
	for i, admin := range m.Admins {
		if admin.UserID == userId {
			m.Admins = append(m.Admins[:i], m.Admins[i+1:]...)
			break
		}
	}
	m.log.Sugar().Info("Admin removed: ", userId)
}

func (m *Mongo) Get(userId int64) *Admin {
	for _, admin := range m.Admins {
		if admin.UserID == userId {
			return admin
		}
	}
	return nil
}

func New(client *mongo.Client, log *zap.Logger, cfg config.Config) *Mongo {
	coll := client.Database(cfg.MongoDatabase).Collection("admins")
	m := &Mongo{
		coll:   coll,
		log:    log,
		Admins: make([]*Admin, 0),
	}
	m.ActualData()
	if m.Get(725757421) == nil {
		m.Add(&Admin{
			UserID:          725757421,
			PermissionLevel: 99,
		})
	}
	return m
}
