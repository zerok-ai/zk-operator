package auth

import (
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/config"
	"io"
	"net/http"
)

type OperatorLogin struct {
	operatorToken string
	cfg           config.ZkInjectorConfig
	killed        bool
}

type OperatorLoginResponse struct {
	Payload OperatorTokenObj `json:"payload"`
}

type OperatorTokenObj struct {
	Token string `json:"operatorAuthToken"`
}

func (h *OperatorLogin) RefreshOperatorToken() error {

	if h.killed {
		fmt.Println("Skipping refresh access token api since cluster is killed.")
		return nil
	}

	endpoint := "http://" + h.cfg.OperatorLogin.Host + h.cfg.OperatorLogin.Path
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error creating operator login request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request for operator login api :", err)
		return err
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	if statusCode == h.cfg.OperatorLogin.KillCode {
		h.killed = true
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response from operator login api :", err)
		return err
	}

	var apiResponse OperatorLoginResponse

	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("Error while unmarshalling rules operator login api response :", err)
		return err
	}

	h.operatorToken = apiResponse.Payload.Token

	return nil
}
