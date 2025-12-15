package steam_accounts

import (
	"bytes"
	"encoding/json"
	"errors"
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
	resp, err := A.ApiRequest("/steam/"+accountId+"/stop", http.MethodPost, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		A.Log.Error("failed to stop farm"+accountId, zap.Any("response", resp))
		return errors.New("failed to stop farm")
	}
	return nil
}

func (A SteamAPI) StartFarming(accountId string) error {
	resp, err := A.ApiRequest("/steam/"+accountId+"/start", http.MethodPost, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		A.Log.Error("failed to start farm"+accountId, zap.Any("response", resp))
		return errors.New("failed to start farm")
	}
	return nil
}

func (A SteamAPI) DeleteAccount(accountId string) error {
	resp, err := A.ApiRequest("/steam/"+accountId, http.MethodDelete, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		A.Log.Error("failed to delete steam account"+accountId, zap.Any("response", resp))
		return errors.New("failed to delete steam account")
	}
	return nil
}

func (A SteamAPI) UpdateStatus(accountId string, state PersonaState) error {
	resp, err := A.ApiRequest("/steam/"+accountId+"/status/"+strconv.FormatInt(int64(state), 10), http.MethodPut, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		A.Log.Error("failed to update status"+accountId, zap.Int("status", int(state)))
		return errors.New("failed to update status")
	}
	return nil
}

func (A SteamAPI) UpdateGames(accountId string, gameIds []any) error {
	bodyBytes, err := json.Marshal(gameIds)
	if err != nil {
		return err
	}
	resp, err := A.ApiRequest("/steam/"+accountId+"/games", http.MethodPut, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		A.Log.Error("failed to update games"+accountId, zap.Any("games", gameIds), zap.Any("response", resp))
		return errors.New("failed to update games")
	}
	return nil
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
