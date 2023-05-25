package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/config"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"io"
	"net/http"
)

type OperatorLogin struct {
	operatorToken  string
	operatorConfig config.OperatorLoginConfig
	killed         bool
}

type OperatorLoginResponse struct {
	Payload OperatorTokenObj `json:"payload"`
}

type OperatorTokenObj struct {
	Token string `json:"operatorAuthToken"`
}

type OperatorLoginRequest struct {
	ClusterKey string `json:"clusterKey"`
}

func CreateOperatorLogin(config config.OperatorLoginConfig) *OperatorLogin {
	opLogin := OperatorLogin{operatorConfig: config, killed: false, operatorToken: ""}
	return &opLogin
}

func (h *OperatorLogin) GetOperatorToken() string {
	return h.operatorToken
}

func (h *OperatorLogin) isKilled() bool {
	return h.killed
}

func (h *OperatorLogin) RefreshOperatorToken() error {

	if h.killed {
		fmt.Println("Skipping refresh access token api since cluster is killed.")
		return nil
	}

	endpoint := "http://" + h.operatorConfig.Host + h.operatorConfig.Path

	clusterKey, err := utils.GetSecretValue(h.operatorConfig.ClusterKeyNamespace, h.operatorConfig.ClusterKey, h.operatorConfig.ClusterKeyData)

	if err != nil {
		fmt.Println("Error while getting cluster key from secrets :", err)
		return err
	}

	requestPayload := OperatorLoginRequest{ClusterKey: clusterKey}

	data, err := json.Marshal(requestPayload)

	if err != nil {
		fmt.Println("Error while creating payload for operator login request:", err)
		return err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(data))

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

	if statusCode == h.operatorConfig.KillCode {
		h.killed = true
		//TODO: Make other changes like stop instrumentation. Maybe in this case we can stop the timer for sync rules and update Orchestration etc.
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
