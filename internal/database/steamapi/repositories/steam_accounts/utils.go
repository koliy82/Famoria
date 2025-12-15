package steam_accounts

import (
	"bytes"
	"io"
	"net/http"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"go.uber.org/zap"
)

func (A SteamAPI) ApiRequest(shortURL string, method string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, A.URL+shortURL, body)
	if err != nil {
		A.Log.Fatal("Failed create steam api request", zap.Error(err))
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+A.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	return A.Client.Do(req)
}

type qrWriteCloser struct {
	*bytes.Buffer
}

func (q qrWriteCloser) Close() error {
	return nil
}

func (A SteamAPI) GenerateQRCode(telegramID int64) ([]byte, error) {
	url, err := A.CreateAccount(telegramID)
	if err != nil {
		A.Log.Error("failed to create qr code url", zap.Int64("user_id", telegramID), zap.Error(err))
		return nil, err
	}
	qrc, err := qrcode.New(url)
	if err != nil {
		A.Log.Error("create qrcode failed: %v\n", zap.Error(err))
		return nil, err
	}
	buf := &bytes.Buffer{}
	wc := qrWriteCloser{buf}

	writer := standard.NewWithWriter(wc,
		standard.WithLogoImageFilePNG("resources/images/qrlogo.png"),
	)

	defer writer.Close()
	if err = qrc.Save(writer); err != nil {
		A.Log.Error("save qrcode failed: %v\n", zap.Error(err))
		return nil, err
	}
	return buf.Bytes(), nil
}
