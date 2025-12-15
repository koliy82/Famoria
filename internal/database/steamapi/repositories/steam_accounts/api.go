package steam_accounts

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"
)

type SteamAPI struct {
	URL    string
	ApiKey string
	Log    *zap.Logger
	Client *http.Client
}

func (A SteamAPI) CreateAccount(telegramId int64) (string, error) {
	//TODO implement me
	return "https://www.youtube.com/", nil
}

func (A SteamAPI) StopFarming(accountId string) error {
	//TODO implement me
	panic("implement me")
}

func (A SteamAPI) StartFarming(accountId string) error {
	//TODO implement me
	panic("implement me")
}

func (A SteamAPI) DeleteAccount(accountId string) error {
	//TODO implement me
	panic("implement me")
}

func (A SteamAPI) UpdateStatus(accountId string, state PersonaState) error {
	//TODO implement me
	panic("implement me")
}

func (A SteamAPI) UpdateGames(accountId string, gameIds []uint32) error {
	//TODO implement me
	panic("implement me")
}

func (A SteamAPI) FindByUserID(id int64) ([]*SteamAccount, error) {
	resp, err := A.ApiRequest("/steam/"+strconv.FormatInt(id, 10), http.MethodGet, nil)
	var steamAccounts []*SteamAccount
	if err != nil {
		return steamAccounts, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		A.Log.Fatal("Failed to get response body", zap.Error(err))
		return steamAccounts, err
	}

	err = json.Unmarshal(body, &steamAccounts)
	if err != nil {
		A.Log.Fatal("Failed unmarshal json accounts", zap.Error(err))
		return steamAccounts, err
	}

	return steamAccounts, nil
}
