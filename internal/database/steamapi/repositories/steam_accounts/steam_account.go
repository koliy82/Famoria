package steam_accounts

type Repository interface {
	FindByUserID(id int64) ([]*SteamAccount, error)
	UpdateGames(accountId string, gameIds []uint32) error
	UpdateStatus(accountId string, state PersonaState) error
	DeleteAccount(accountId string) error
	StartFarming(accountId string) error
	StopFarming(accountId string) error
	CreateAccount(telegramId int64) (string, error)
}
