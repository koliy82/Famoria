package steam_accounts

type Repository interface {
	FindByUserID(id int64) ([]*SteamAccount, error)
	UpdateGames(accountId string, gameIds []any) error
	UpdateStatus(accountId string, state PersonaState) error
	DeleteAccount(accountId string) error
	StartFarming(accountId string) error
	StopFarming(accountId string) error
	CreateWithQR(telegramId int64) (string, error)
	BasicAuth(telegramId int64, login string, password string) error
	Confirm2FA(telegramId int64, code string) error
}
