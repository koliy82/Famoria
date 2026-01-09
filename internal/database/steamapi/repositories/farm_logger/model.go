package farm_logger

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FarmLog struct {
	OID        primitive.ObjectID `bson:"_id"`
	SteamID    string             `bson:"steam_id"`
	TelegramID int64              `bson:"telegram_id"`
	SteamName  *string            `bson:"steam_name"`
	State      SessionState       `bson:"state"`
	Reason     LogReason          `bson:"reason"`
	Date       time.Time          `bson:"date"`
}

func (p *FarmLog) SteamUsername() string {
	if p.SteamName == nil {
		return p.SteamID
	}
	return *p.SteamName
}

type SessionState int

const (
	Unknown SessionState = iota
	Active
	NeedAuth
	Stopped
	Deleted
	TryAnotherCMS
)

func (s SessionState) String() string {
	return [...]string{"Неизвестно", "Активна", "Нужна авторизация", "Остановленна", "Удалена", "Переподключение"}[s]
}

type LogReason int

const (
	GamesSend LogReason = iota
	UserStop
	UserDelete
	AuthError
	ConnectionError
	UnknownError
	LoggedInElsewhere
	TryAnotherCM
)

func (s LogReason) String() string {
	return [...]string{"Запуск игр", "Остановка пользователем", "Удаление аккаунта", "Ошибка авторизации", "Ошибка подключения", "Неизвестная ощибка", "Игра пользователя", "Отключение от сервера Steam"}[s]
}
