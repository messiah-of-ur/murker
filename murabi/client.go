package murabi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type FinishRequest struct {
	GameID string `json:"gameID"`
	Winner int    `json:"winner"`
}

type MurabiClient struct {
	client     *http.Client
	murabiAddr string
}

func NewMurabiClient(murabiAddr string) *MurabiClient {
	return &MurabiClient{
		client:     &http.Client{},
		murabiAddr: murabiAddr,
	}
}

func (m *MurabiClient) FinishGame(finishReq *FinishRequest) error {
	endpoint := "http://" + m.murabiAddr + "/murabi/v1/game/finish"
	reqBodyAsStr, err := json.Marshal(finishReq)

	if err != nil {
		return fmt.Errorf("Error marshalling payload: %s", err.Error())
	}

	payload := strings.NewReader((string(reqBodyAsStr)))
	req, err := http.NewRequest("POST", endpoint, payload)

	if err != nil {
		return fmt.Errorf("Error creating request: %s", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending request: %s", err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Server responded with an error http code: %d", res.StatusCode)
	}

	return nil
}
